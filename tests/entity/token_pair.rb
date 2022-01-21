# frozen_string_literal: true

# TokenPair is a class to represent a pair of tokens
class TokenPair
  include Jsonizable

  attr_accessor :access_token, :refresh_token, :token_type, :expires_in

  def initialize(access_token, refresh_token, token_type, expires_in)
    @access_token = access_token
    @refresh_token = refresh_token
    @token_type = token_type
    @expires_in = expires_in
  end

  def to_hash
    {
      access_token: @access_token,
      refresh_token: @refresh_token,
      token_type: @token_type,
      expires_in: @expires_in
    }
  end

  class << self
    def empty_token_pairs
      TokenPair.new '', '', 'Bearer', 0
    end

    def from_hash(hash)
      TokenPair.new hash[:access_token], hash[:refresh_token], hash[:token_type], hash[:expires_in]
    end
  end
end
