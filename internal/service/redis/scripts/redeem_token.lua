local tokenKey = KEYS[1]
local resultKey = KEYS[2]
local language = ARGV[1]
local expectedTokenValue = ARGV[2]

local raw = redis.call('GET', tokenKey)
if not raw then return 'not_found' end
if raw ~= expectedTokenValue then return 'retry' end

local separator = string.find(raw, ':', 1, true)
if not separator then return 'invalid_state' end
if string.sub(raw, 1, separator - 1) ~= language then return 'wrong_language' end

redis.call('DEL', tokenKey, resultKey)
return 'redeemed'
