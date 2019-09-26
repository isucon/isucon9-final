require "sinatra/base"

class App < Sinatra::Base
  set :bind, "0.0.0.0"

  get '/api' do
    "Hello World"
  end
end
