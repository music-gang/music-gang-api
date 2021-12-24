# frozen_string_literal: true

# AuthService is a class to represent a service test auth api flows
class AuthService < ServiceHTTP
  # register a new user
  # @param [URI] url
  # @param [User] user
  # @return [TokenPair]
  def register(user)
    raise ArgumentError, 'Expected User' unless user.is_a? User

    url = URI("#{base_url}/auth/register")

    http = Net::HTTP.new(url.host, url.port)

    request = Net::HTTP::Post.new(url)
    request['Content-Type'] = 'application/json'
    request.body = { name: user.name,
                     email: user.email,
                     password: user.password,
                     confirm_password: user.password }.to_json
    # @type [Net::HTTPResponse]
    response = http.request(request)
    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    TokenPair.from_hash JSON.parse(response.body)
  end
end
