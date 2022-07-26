# frozen_string_literal: true

# ContractService
class ContractService < ServiceHTTP
  # Create a new contract.
  # @param contract [Contract]
  # @return [Contract]
  def create(contract)
    url = URI("#{base_url}/contract")

    http = Net::HTTP.new url.host, url.port

    request = Net::HTTP::Post.new url

    request.content_type = 'application/json'
    request.body = contract.to_json

    # @type [Net::HTTPResponse]
    response = http.request request
    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    Contract.from_hash JSON.parse(response.body, symbolize_names: true)
  end
end
