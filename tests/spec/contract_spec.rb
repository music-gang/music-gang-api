# frozen_string_literal: true

require_relative '../services/lib'
require_relative 'util'

container = service_container

# @param user [User]
def success_create_contract(user, stateful = false)
  contract = Contract.new(
    name: 'Contract 1',
    description: 'Contract 1 description' + (stateful ? ' (stateful)' : ''),
    user_id: user.id,
    visibility: 'public',
    stateful: stateful,
    max_fuel: 100
  )
  service_container.contract_service.create(access_token: user.token_pairs.access_token, contract: contract)
end

def success_create_revision(user, contract)
  service_container.contract_service.make_revision access_token: user.token_pairs.access_token, contract: contract, revision: (Revision.new version: 'Anchorage', notes: 'New revision ready to be executed!', max_fuel: 3000, contract_id: contract.id)
end

describe 'Flow Contract:' do
  user = success_register_user

  describe 'create a new contract' do
    context 'given correct data' do
      it 'returns the contract created' do
        contract = Contract.new name: "Contract #{Faker::Lorem.word}", description: Faker::Lorem.paragraph, user_id: user.id, visibility: 'public', max_fuel: 5000

        container.contract_service.create contract: contract, access_token: user.token_pairs.access_token
      end
    end

    context 'given incorrect data' do
      context 'for example an empty name' do
        it 'returns an error' do
          contract = Contract.new name: '', description: Faker::Lorem.paragraph, user_id: user.id, visibility: 'public', max_fuel: 5000
          container.contract_service.create contract: contract, access_token: user.token_pairs.access_token
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end

      context 'for example an empty user' do
        it 'returns an error' do
          contract = Contract.new name: "Contract #{Faker::Lorem.word}", description: Faker::Lorem.paragraph, user_id: nil, visibility: 'public', max_fuel: 5000
          container.contract_service.create contract: contract, access_token: user.token_pairs.access_token
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end

      context 'for example an empty max_fuel' do
        it 'returns an error' do
          contract = Contract.new name: "Contract #{Faker::Lorem.word}", description: Faker::Lorem.paragraph, user_id: user.id, visibility: 'public', max_fuel: nil
          container.contract_service.create contract: contract, access_token: user.token_pairs.access_token
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end

      context 'for example an empty visibility' do
        it 'returns an error' do
          contract = Contract.new name: "Contract #{Faker::Lorem.word}", description: Faker::Lorem.paragraph, user_id: user.id, visibility: nil, max_fuel: 5000
          container.contract_service.create contract: contract, access_token: user.token_pairs.access_token
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end

      context 'for example an invalid visibility' do
        it 'returns an error' do
          contract = Contract.new name: "Contract #{Faker::Lorem.word}", description: Faker::Lorem.paragraph, user_id: user.id, visibility: 'invalid-visibility', max_fuel: 5000
          container.contract_service.create contract: contract, access_token: user.token_pairs.access_token
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end
    end
  end

  describe 'update a contract' do
    user = success_register_user

    context 'given correct data' do
      it 'returns the contract updated' do
        contract = success_create_contract user

        contract.name = "New name for contract #{contract.id}"
        contract.description = "New description for contract #{contract.id}"
        contract.max_fuel = 150

        updated_contract = container.contract_service.update access_token: user.token_pairs.access_token, contract: contract

        expect(updated_contract.name).to eq contract.name
        expect(updated_contract.description).to eq contract.description
        expect(updated_contract.max_fuel).to eq contract.max_fuel
      end
    end

    context 'given invalid contract id' do
      it 'returns an error' do
        contract = success_create_contract user

        contract.name = "New name for contract #{contract.id}"
        contract.description = "New description for contract #{contract.id}"
        contract.max_fuel = 150
        contract.id = 0

        container.contract_service.update access_token: user.token_pairs.access_token, contract: contract
      rescue ServiceError
        'error correctley raised'
      else
        raise 'error not raised'
      end
    end
  end

  describe 'retrieve a contract' do
    context 'given a valid contract id' do
      it 'returns the contract' do
        contract = success_create_contract user

        fetched_contract = container.contract_service.get access_token: user.token_pairs.access_token, contract_id: contract.id

        expect(fetched_contract.name).to eq contract.name
      end
    end

    context 'given an invalid contract id' do
      it 'returns an error' do
        container.contract_service.get access_token: user.token_pairs.access_token, contract_id: 0

      rescue ServiceError
        'error correctley raised'
      else
        raise 'error not raised'
      end
    end
  end

  describe 'make a new revision' do
    context 'given a valid revision' do
      it 'returns the revision' do
        contract = success_create_contract user
        revision = Revision.new version: 'Anchorage', notes: 'New revision!', max_fuel: 3000, contract_id: contract.id

        rev1 = container.contract_service.make_revision access_token: user.token_pairs.access_token, contract: contract, revision: revision
        rev2 = container.contract_service.make_revision access_token: user.token_pairs.access_token, contract: contract, revision: revision

        expect(rev1.version).to eq rev2.version
        expect(rev1.rev).not_to eq rev2.rev
      end
    end

    context 'given a invalid revision' do
      context 'for example an empty version' do
        it 'returns an error' do
          contract = success_create_contract user
          revision = Revision.new notes: 'New revision!', max_fuel: 3000, contract_id: contract.id

          container.contract_service.make_revision access_token: user.token_pairs.access_token, contract: contract, revision: revision
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end
    end
  end

  describe 'execute a contract' do
    context 'stateless contract' do
      contract = success_create_contract user
      revision = success_create_revision user, contract

      context 'given a valid contract id and rev number' do
        it 'returns the contract result' do
          result = container.contract_service.execute access_token: user.token_pairs.access_token, contract_id: contract.id, rev: revision.rev
          expect(result).to eq 'Hello World!'
        end
      end
    end

    context 'stateful contract' do
      contract_stateful = success_create_contract user, true
      revision_stateful = success_create_revision user, contract_stateful

      context 'given a valid contract id and rev number' do
        it 'returns the contract result' do
          res1 = container.contract_service.execute access_token: user.token_pairs.access_token, contract_id: contract_stateful.id, rev: revision_stateful.rev
          expect(res1).to eq '1'
          res2 = container.contract_service.execute access_token: user.token_pairs.access_token, contract_id: contract_stateful.id, rev: revision_stateful.rev
          expect(res2).to eq '2'
          res3 = container.contract_service.execute access_token: user.token_pairs.access_token, contract_id: contract_stateful.id, rev: revision_stateful.rev
          expect(res3).to eq '3'
        end
      end
    end
  end
end
