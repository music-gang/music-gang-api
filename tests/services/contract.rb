# frozen_string_literal: true

# ContractService
class ContractService < ServiceHTTP
  # Create a new contract.
  # @param access_token [String]
  # @param contract [Contract]
  # @return [Contract]
  def create(access_token: nil, contract: nil)
    raise 'access_token is required' if access_token.nil?

    url = URI("#{base_url}/contract")

    http = Net::HTTP.new url.host, url.port

    request = Net::HTTP::Post.new url

    request.content_type = 'application/json'
    # add Request header
    request.add_field 'Authorization', "Bearer #{access_token}"
    request.body = contract.to_json

    # @type [Net::HTTPResponse]
    response = http.request request
    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    Contract.from_hash JSON.parse(response.body, symbolize_names: true)[:contract]
  end

  # Execute a contract
  # @param access_token [String]
  # @param contract_id [Integer]
  # @param rev [String]
  def execute(access_token: nil, contract_id: nil, rev: nil)
    raise 'access_token is required' if access_token.nil?
    raise 'contract_id is required' if contract_id.nil?

    url = URI("#{base_url}/contract/#{contract_id}/call#{rev.nil? ? '' : "/#{rev}"}")

    http = Net::HTTP.new url.host, url.port

    request = Net::HTTP::Post.new url

    request.content_type = 'application/json'
    # add Request header
    request.add_field 'Authorization', "Bearer #{access_token}"

    # @type [Net::HTTPResponse]
    response = http.request request
    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    JSON.parse(response.body, symbolize_names: true)[:result]
  end

  # Retrieve a contract
  # @param access_token [String]
  # @param contract_id [Integer]
  # @return [Contract]
  def get(access_token: nil, contract_id: nil)
    raise 'access_token is required' if access_token.nil?
    raise 'contract_id is required' if contract_id.nil?

    url = URI("#{base_url}/contract/#{contract_id}")

    http = Net::HTTP.new url.host, url.port

    request = Net::HTTP::Get.new url

    request.content_type = 'application/json'
    # add Request header
    request.add_field 'Authorization', "Bearer #{access_token}"

    # @type [Net::HTTPResponse]
    response = http.request request
    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    Contract.from_hash JSON.parse(response.body, symbolize_names: true)[:contract]
  end

  # Make a new revision of a contract
  # @param access_token [String]
  # @param contract [Contract]
  # @param revision [Revision]
  # @return [Revision]
  def make_revision(access_token: nil, contract: nil, revision: nil)
    raise 'access_token is required' if access_token.nil?
    raise 'revision is required' if revision.nil?
    raise 'contract_id is required' if revision.contract_id.nil?

    url = URI("#{base_url}/contract/#{revision.contract_id}/revision")

    http = Net::HTTP.new url.host, url.port

    # retrive file for multipart
    revision_compiled = File.open(contract.stateful ? './examples/stateful_revision.js' : './examples/simple_revision.js', 'rb')

    # @type [Net::HTTP::Post]
    request = Net::HTTP::Post::Multipart.new url, { 'revision' => revision.to_json, 'compiled_revision' => UploadIO.new(revision_compiled, 'application/javascript', 'simple_revision.js') }

    # add Request header
    request.add_field 'Authorization', "Bearer #{access_token}"

    # @type [Net::HTTPResponse]
    response = http.request request
    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    revision_compiled.close

    Revision.from_hash JSON.parse(response.body, symbolize_names: true)[:revision]
  end

  # Update a contract
  # @param access_token [String]
  # @param contract [Contract]
  def update(access_token: nil, contract: nil)
    raise 'access_token is required' if access_token.nil?
    raise 'contract is required' if contract.nil?

    url = URI("#{base_url}/contract/#{contract.id}")

    http = Net::HTTP.new url.host, url.port

    request = Net::HTTP::Put.new url

    request.content_type = 'application/json'
    # add Request header
    request.add_field 'Authorization', "Bearer #{access_token}"
    request.body = { name: contract.name, description: contract.description, max_fuel: contract.max_fuel }.to_json

    # @type [Net::HTTPResponse]
    response = http.request request
    raise ServiceError.new response.body, response.code unless response.is_a? Net::HTTPSuccess

    Contract.from_hash JSON.parse(response.body, symbolize_names: true)[:contract]
  end
end
