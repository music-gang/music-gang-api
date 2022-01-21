# frozen_string_literal: true

require 'faker'

require_relative '../services/lib'

container = service_container

name = Faker::Name.name
email = Faker::Internet.email name: name, separators: '.'
password = 'Password123!'

user = User.new name, email, password

describe 'Flow Auth:' do
  describe 'register a new user' do
    context 'given correct data' do
      it 'returns token pairs' do
        container.auth_service.register user: user
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
          end
        end
      end
    end
  end

  describe 'login a user' do
    context 'given correct data' do
      it 'returns token pairs' do
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
          container.auth_service.login email: user.email, password: 'non-valid-password'
        rescue ServiceError
          'error correctley raised'
        else
        end
      end

      context 'for example an empty password' do
        it 'returns an error' do
          container.auth_service.login email: user.email, password: ''
        rescue ServiceError
          'error correctley raised'
        end
      end

      context 'for example a non valid email' do
        it 'returns an error' do
          container.auth_service.login email: 'incorrect-email', password: user.password
        rescue ServiceError
          'error correctley raised'
        end
      end

      context 'for example an empty email' do
        it 'returns an error' do
          container.auth_service.login email: '', password: user.password
        rescue ServiceError
          'error correctley raised'
        end
      end

      context 'for example a not existing email' do
        it 'returns an error' do
          container.auth_service.login email: Faker::Internet.email, password: user.password
        rescue ServiceError
          'error correctley raised'
        end
      end
    end
  end

  describe 'refresh a token pairs' do
    old_pairs = nil

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
        end
      end

      context 'for example using an old refresh token' do
        it 'returns an error' do
          container.auth_service.refresh refresh_token: old_pairs.refresh_token
        rescue ServiceError
          'error correctley raised'
        end
      end
    end
  end

  describe 'logout a user' do
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
        end
      end

      context 'for example an incorrect access token' do
        it 'returns an error' do
          empty_token_pairs = TokenPair.empty_token_pairs
          empty_token_pairs.access_token = 'non-valid-token'
          container.auth_service.logout pairs: empty_token_pairs
        rescue ServiceError
          'error correctley raised'
        end
      end
    end
  end
end
