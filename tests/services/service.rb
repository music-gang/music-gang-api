# frozen_string_literal: true

# ServiceContainer contains all the services
class ServiceContainer
  attr_accessor :services

  def initialize
    @services = {}
  end

  # @return [AuthService]
  def auth_service
    raise ServiceNotFound, 'auth' unless @services.key? :auth

    @services[:auth]
  end

  # @return [FuelService]
  def fuel_service
    raise ServiceNotFound, 'fuel' unless @services.key? :fuel

    @services[:fuel]
  end
end

# ServiceNotFound is raised when a service is not found
class ServiceNotFound < StandardError
  def initialize(service_name)
    super "Service #{service_name} not found"
  end
end
