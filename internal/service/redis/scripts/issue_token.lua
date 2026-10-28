local captchaKey = KEYS[1]
local resultKey = KEYS[2]
local candidateKey = KEYS[3]
local priorTokenKey = KEYS[4]

local passwordCorrect = ARGV[1] == '1'
local language = ARGV[2]
local candidate = ARGV[3]
local tokenTTL = tonumber(ARGV[4])
local captchaID = ARGV[5]
local expectedResult = ARGV[6]
local expectedPriorToken = ARGV[7]

local prior = redis.call('GET', resultKey)
if prior then
  if not passwordCorrect then return {'invalid_password', ''} end
  if prior ~= expectedResult then return {'retry', prior} end

  local separator = string.find(prior, ':', 1, true)
  if not separator then return {'invalid_state', ''} end
  local priorLanguage = string.sub(prior, 1, separator - 1)
  local priorToken = string.sub(prior, separator + 1)
  if priorLanguage ~= language then return {'unsupported_language', ''} end
  if priorToken ~= expectedPriorToken then return {'invalid_state', ''} end
  if redis.call('EXISTS', priorTokenKey) == 1 then
    return {'reused', priorToken}
  end

  redis.call('DEL', resultKey)
  return {'not_found', ''}
end

if expectedResult ~= '' then return {'retry', ''} end
local raw = redis.call('GET', captchaKey)
if not raw then return {'not_found', ''} end
local state = cjson.decode(raw)
if not state.solved then return {'not_solved', ''} end

if not passwordCorrect then
  if redis.call('PTTL', captchaKey) <= 0 then return {'invalid_state', ''} end
  state.triesLeft = state.triesLeft - 1
  if state.triesLeft <= 0 then
    redis.call('DEL', captchaKey)
    return {'no_tries', ''}
  end
  redis.call('SET', captchaKey, cjson.encode(state), 'KEEPTTL')
  return {'invalid_password', ''}
end

if not tokenTTL or tokenTTL <= 0 then return {'invalid_state', ''} end
if redis.call('EXISTS', candidateKey) == 1 then return {'collision', ''} end
redis.call('PSETEX', candidateKey, tokenTTL, language .. ':' .. captchaID)
redis.call('PSETEX', resultKey, tokenTTL, language .. ':' .. candidate)
redis.call('DEL', captchaKey)
return {'issued', candidate}
