# frozen_string_literal: true

require_relative '../util/lib'
require_relative '../entity/lib'

require_relative 'service'

require_relative 'auth'

SCHEMA = ENV['MUSICGANG_SERVICE_SCHEMA'] || 'http'
URL = ENV['MUSICGANG_SERVICE_URL'] || 'localhost:8888/v1'

# @return [ServiceContainer]
def service_container
  @container = ServiceContainer.new
  @container.services[:auth] = AuthService.new SCHEMA, URL
  @container
end

# ServiceError
class ServiceError < StandardError
  attr_reader :http_code, :error_code

  def initialize(message, http_code)
    @http_code = http_code
    super message
  end
end
