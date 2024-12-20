              -- Bitwise operation implementations for Lua 5.1
              local function band(a, b)
                local result = 0
                local bitval = 1
                while a > 0 and b > 0 do
                  if a % 2 == 1 and b % 2 == 1 then
                    result = result + bitval
                  end
                  bitval = bitval * 2
                  a = math.floor(a/2)
                  b = math.floor(b/2)
                end
                return result
              end

              local function rshift(a, b)
                return math.floor(a / (2^b))
              end

              local function lshift(a, b)
                return a * (2^b)
              end

              -- Helper function to read varint
              function read_varint(bytes, offset)
                local result = 0
                local shift = 0
                local pos = offset or 1

                while pos <= #bytes do
                  local byte = string.byte(bytes, pos)
                  result = result + lshift(band(byte, 0x7F), shift)
                  pos = pos + 1

                  if band(byte, 0x80) == 0 then
                    break
                  end
                  shift = shift + 7
                end

                return result, pos - offset
              end

              -- Helper function to read length-delimited field
              function read_length_delimited(bytes, offset)
                local length, varint_size = read_varint(bytes, offset)
                local start = offset + varint_size
                local value = string.sub(bytes, start, start + length - 1)
                return value, varint_size + length
              end

              -- Function to decode map entry (key-value pair)
              function decode_map_entry(bytes)
                local pos = 1
                local key, value

                while pos <= #bytes do
                  -- Read field tag
                  local tag = string.byte(bytes, pos)
                  local field_num = rshift(tag, 3)
                  local wire_type = band(tag, 0x07)
                  pos = pos + 1

                  -- field 1 is key (string)
                  -- field 2 is value (string)
                  if wire_type == 2 then -- length-delimited
                    local field_value, bytes_read = read_length_delimited(bytes, pos)
                    pos = pos + bytes_read

                    if field_num == 1 then
                      key = field_value
                    elseif field_num == 2 then
                      value = field_value
                    end
                  else
                    break -- Unknown wire type
                  end
                end

                return key, value
              end

              -- Main function to decode the entire message
              function decode_overrides(bytes)
                local pos = 1
                local result = {}

                while pos <= #bytes do
                  local tag = string.byte(bytes, pos)
                  local field_num = rshift(tag, 3)
                  local wire_type = band(tag, 0x07)
                  pos = pos + 1

                  -- field 1 is routing_overrides (map)
                  if field_num == 1 and wire_type == 2 then
                    local map_entry, bytes_read = read_length_delimited(bytes, pos)
                    local key, value = decode_map_entry(map_entry)

                    if key and value then
                      result[key] = value
                    end

                    pos = pos + bytes_read
                  else
                    break -- Unknown field
                  end
                end

                return result
              end

              -- Base64 decode function
              function base64_decode(data)
                if not data then return nil end

                local b = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/'
                data = string.gsub(data, '[^'..b..'=]', '')
                return (data:gsub('.', function(x)
                  if (x == '=') then return '' end
                  local r,f='',b:find(x)-1
                  for i=6,1,-1 do r=r..(f%2^i-f%2^(i-1)>0 and '1' or '0') end
                  return r;
                end):gsub('%d%d%d?%d?%d?%d?%d?%d?', function(x)
                  if (#x ~= 8) then return '' end
                  local c=0
                  for i=1,8 do c=c+(x:sub(i,i)=='1' and 2^(8-i) or 0) end
                  return string.char(c)
                end))
              end

              function split_baggage(baggage_str)
                local result = {}
                -- Split on commas that are not within quotes
                for item in baggage_str:gmatch("[^,]+") do
                  -- Trim whitespace
                  item = item:match("^%s*(.-)%s*$")

                  -- Split key=value
                  local key, value = item:match("([^=]+)=(.*)")
                  if key and value then
                    -- Trim quotes if present
                    value = value:gsub("^\"(.-)\"$", "%1")
                    result[key:match("^%s*(.-)%s*$")] = value
                  end
                end
                return result
              end

              function handle_baggage(request_handle)
                local baggage = request_handle:headers():get("baggage")

                if not baggage then
                  request_handle:logInfo("No baggage found")
                  return
                end
                request_handle:logInfo("baggage: " .. baggage)
                local baggage_items = split_baggage(baggage)
                local overrides_value = baggage_items["overrides"]
                if not overrides_value then
                  request_handle:logInfo("No overrides found")
                  return
                end
                request_handle:logInfo("Found overrides value: " .. overrides_value)
                return overrides_value
              end

              function envoy_on_request(request_handle)
                local proto_header = handle_baggage(request_handle)
                if not proto_header then
                  request_handle:logInfo("No proto header found")
                  return
                end
                local decoded_base64 = base64_decode(proto_header)

                local overrides = decode_overrides(decoded_base64)

                -- Print individual results
                for k, v in pairs(overrides) do
                  request_handle:logInfo(k .. ": " .. v)
                  if not request_handle:headers():get("x-" .. k) then
                    request_handle:headers():add("x-" .. k , v)
                  end
                end
              end
