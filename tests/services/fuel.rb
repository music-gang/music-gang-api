# frozen_string_literal: true

# FuelService is a class to access fuel api
class FuelService < ServiceHTTP
  # Fuel stats
  def stats
    url = URI.parse("#{base_url}/vm/stats")

    http = Net::HTTP.new url.host, url.port

    request = Net::HTTP::Get.new url

    response = http.request request

    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    data = JSON.parse(response.body, symbolize_names: true)

    raise 'stats key not found' unless data.key? :stats

    FuelStat.from_hash data[:stats]
  end
end
