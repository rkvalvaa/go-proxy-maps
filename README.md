# go-proxy-maps
Go service that act as a simple proxy between your frontend and Google Maps Javascript API, so that you can hide your API key

It does have rate limiting and simple cache implemented in the code, which can be adjusted based on your needs.
It does currently not have CORS implemented to restrict it to given frontend domains, but this can easily be implemented by using the github.com/rs/cors package
