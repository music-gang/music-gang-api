# frozen_string_literal: true

# ServiceContainer contains all the services
class ServiceContainer
  attr_accessor :services

  def initialize
    @services = {}
  end

  # @return [AuthService]
  def auth_service
    @services[:auth]
  end
end
