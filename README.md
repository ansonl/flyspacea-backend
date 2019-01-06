Fly Space-A Photo Processor Server
===================

![Fly Space-A logo](https://avatars1.githubusercontent.com/u/38817545?s=200&v=4)

This the backend server running Fly Space-A. The backend downloads flight schedule photos, processes flight schedules into text data, and provides the flight schedules to client applications over REST API. This code is fully functional. Please see the [technical implementation](https://docs.google.com/presentation/d/1cnS_nTL6xhL5PEHFro7jvDSuHAccr8eSBFV26KIfrzE/edit?usp=sharing) slides for more detailed information on how photos are processed into text.

<iframe src="https://docs.google.com/presentation/d/e/2PACX-1vTI4-2m6gRjPTHvhui8BpOhT46IQ1-k-eYYPKgiapNFGLrbsLqZzNdUYikU6xa2_3XQvzTJ78bIToki/embed?start=false&loop=false&delayms=3000" frameborder="0" width="960" height="749" allowfullscreen="true" mozallowfullscreen="true" webkitallowfullscreen="true"></iframe>

How to use
-------------

1. Install Go

2. `go get https://github.com/ansonl/flyspacea-backend`

3. Paste your Facebook Graph API Access Token into `constants.go`.

4. Set $DATABASE_URL to your PostgreSQL database URL.

3. `go install spacea`

4. `spacea`

Credits
-------------

[jbowtie](https://github.com/jbowtie) fork of [gokogiri](https://github.com/jbowtie/gokogiri)

[latlng](github.com/bradfitz/latlong) by bradfitz

[pq](github.com/lib/pq) - Golang PostgreSQL driver

[Fuzzy](https://github.com/sajari/fuzzy) by Sajari

[goprocinfo](https://github.com/c9s/goprocinfo) by c9s