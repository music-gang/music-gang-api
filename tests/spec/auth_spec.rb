# frozen_string_literal: true

require 'faker'

require_relative '../services/lib'
require_relative 'util'

container = service_container

# @return [User]
def success_register_user
  user = new_user
  token_pairs = service_container.auth_service.register(user: user)
  fetched_user = service_container.auth_service.user(access_token: token_pairs.access_token)
  user.token_pairs = token_pairs
  user.id = fetched_user.id
  user
end

describe 'Flow Auth:' do
  describe 'register a new user' do
    context 'given correct data' do
      it 'returns token pairs' do
        container.auth_service.register user: new_user
      end
    end

    context 'given incorrect data' do
      context 'for example an empty name' do
        it 'returns an error' do
          name = ''

          begin
            container.auth_service.register user: User.new(name, email, password)
          rescue ServiceError
            'error correctley raised'
          else
            raise 'error not raised'
          end
        end
      end

      context 'for example an incorrect password' do
        it 'returns an error' do
          name = Faker::Name.name
          email = Faker::Internet.email name: name, separators: '.'
          password = 'non-valid-password'

          begin
            container.auth_service.register user: User.new(name, email, password)
          rescue ServiceError
            'error correctley raised'
          else
            raise 'error not raised'
          end
        end
      end

      context 'for example an incorrect email' do
        it 'returns an error' do
          name = Faker::Name.name
          email = 'incorrect-email'

          begin
            container.auth_service.register user: User.new(name, email, password)
          rescue ServiceError
            'error correctley raised'
          else
            raise 'error not raised'
          end
        end
      end
    end
  end

  describe 'login a user' do
    context 'given correct data' do
      it 'returns token pairs' do
        user = success_register_user

        pairs = container.auth_service.login email: user.email, password: user.password

        expect(pairs.access_token).not_to be_nil
        expect(pairs.refresh_token).not_to be_nil
        expect(pairs.expires_in).not_to be_nil
        expect(pairs.token_type).not_to be_nil

        raise 'Access Token is the same as the refresh token' unless pairs.access_token != pairs.refresh_token

        user.token_pairs = pairs
      end
    end

    context 'given incorrect data' do
      context 'for example an incorrect password' do
        it 'returns an error' do
          user = success_register_user

          container.auth_service.login email: user.email, password: 'non-valid-password'
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end

      context 'for example an empty password' do
        it 'returns an error' do
          user = success_register_user

          container.auth_service.login email: user.email, password: ''
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end

      context 'for example a non valid email' do
        it 'returns an error' do
          container.auth_service.login email: 'incorrect-email', password: 'blabla'
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end

      context 'for example an empty email' do
        it 'returns an error' do
          container.auth_service.login email: '', password: 'blabla'
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end

      context 'for example a not existing email' do
        it 'returns an error' do
          container.auth_service.login email: Faker::Internet.email, password: 'blabla'
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end
    end
  end

  describe 'refresh a token pairs' do
    old_pairs = nil
    user = success_register_user

    context 'given correct data' do
      it 'returns token pairs' do
        pairs = container.auth_service.refresh pairs: user.token_pairs
        expect(pairs.access_token).not_to be_nil
        expect(pairs.access_token).not_to eq(user.token_pairs.access_token)
        expect(pairs.refresh_token).not_to be_nil
        expect(pairs.refresh_token).not_to eq(user.token_pairs.refresh_token)
        user.token_pairs = pairs
        old_pairs = pairs
      end
    end

    context 'given incorrect data' do
      context 'for example an incorrect refresh token' do
        it 'returns an error' do
          container.auth_service.refresh refresh_token: 'non-valid-token'
        rescue ServiceError
          'error correctley raised'
        else
          raise 'error not raised'
        end
      end

      context 'for example using an old refresh token' do
        it 'returns an error' do
          container.auth_service.refresh refresh_token: old_pairs.refresh_token
        rescue ServiceError
          'error correctley raised'
        else
          # no raise here
        end
      end
    end
  end

  describe 'logout a user' do
    user = success_register_user

    context 'given correct data' do
      it 'effectively logs out the user' do
        container.auth_service.logout pairs: user.token_pairs
        user.token_pairs = nil
      end
    end

    context 'given empty data' do
      it 'logout without errors' do
        container.auth_service.logout pairs: TokenPair.empty_token_pairs
      end
    end

    context 'given incorrect data' do
      context 'for example an incorrect refresh token' do
        it 'returns an error' do
          # @type [TokenPair]
          empty_token_pairs = TokenPair.empty_token_pairs
          empty_token_pairs.refresh_token = 'non-valid-token'
          container.auth_service.logout pairs: empty_token_pairs
        rescue ServiceError
          'error correctley raised'
        else
          # no raise here
        end
      end

      context 'for example an incorrect access token' do
        it 'returns an error' do
          empty_token_pairs = TokenPair.empty_token_pairs
          empty_token_pairs.access_token = 'non-valid-token'
          container.auth_service.logout pairs: empty_token_pairs
        rescue ServiceError
          'error correctley raised'
        else
          # not raise here
        end
      end
    end
  end
end
