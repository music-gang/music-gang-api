# frozen_string_literal: true

# AuthService is a class to represent a service test auth api flows
class AuthService < ServiceHTTP
  # login a user
  # @param [String] email
  # @param [String] password
  # @return [TokenPair]
  def login(email: nil, password: nil)
    url = URI("#{base_url}/auth/login")

    http = Net::HTTP.new url.host, url.port

    request = Net::HTTP::Post.new url

    request.content_type = 'application/json'
    request.body = { email: email, password: password }.to_json

    # @type [Net::HTTPResponse]
    response = http.request request
    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    TokenPair.from_hash JSON.parse(response.body, symbolize_names: true)
  end

  # Logut a user
  # @param [TokenPair] pairs
  def logout(pairs: nil)
    raise ArgumentError, 'pairs must be provided' if pairs.nil?

    url = URI("#{base_url}/auth/logout")

    http = Net::HTTP.new url.host, url.port

    request = Net::HTTP::Delete.new url

    request.content_type = 'application/json'
    request.body = pairs.to_json

    # @type [Net::HTTPResponse]
    response = http.request request

    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess
  end

  # Refresh a token given a refresh token or a token pair
  # @param [String] refresh_token
  # @param [TokenPair] pairs
  def refresh(refresh_token: nil, pairs: nil)
    raise ArgumentError, 'refresh_token or pairs must be provided' if refresh_token.nil? && pairs.nil?

    refresh_token = pairs.refresh_token if refresh_token.nil? && !pairs.nil?

    url = URI("#{base_url}/auth/refresh")

    http = Net::HTTP.new url.host, url.port

    request = Net::HTTP::Post.new url

    request.content_type = 'application/json'
    request.body = { refresh_token: refresh_token }.to_json

    # @type [Net::HTTPResponse]
    response = http.request request

    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    TokenPair.from_hash JSON.parse(response.body, symbolize_names: true)
  end

  # register a new user
  # @param [User] user
  # @return [TokenPair]
  def register(user: nil)
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

    TokenPair.from_hash JSON.parse(response.body, symbolize_names: true)
  end
end
