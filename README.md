Fly Space-A Photo Processor Server
===================

![Fly Space-A logo](https://avatars1.githubusercontent.com/u/38817545?s=200&v=4)

This the backend server running Fly Space-A. The backend downloads USAF AMC Space Available flight schedule photos from Facebook, processes flight schedule photos into text data, and provides the flight schedules to client applications over REST API. This code is fully functional. You will need to a free [Facebook Graph](https://developers.facebook.com/) API access token. 

Please see the [technical implementation](https://docs.google.com/presentation/d/1cnS_nTL6xhL5PEHFro7jvDSuHAccr8eSBFV26KIfrzE/edit?usp=sharing) slides also available in `assets` directory for more detailed information on how photos are processed into text.

![highlight fsa](https://raw.githubusercontent.com/ansonl/flyspacea-backend/master-public/assets/fsa_results_highlight.png)

Why is this released? Can I use it for my own projects?
-------------

This backend used to provide the information needed for the free Fly Space-A service, but [*Facebook Graph API **Page Public Content Access***](https://developers.facebook.com/docs/graph-api/reference/page/) was revoked in mid-2018 during tightening of Graph API accesses due/related to the [2018 Cambridge Analytica news](https://en.wikipedia.org/wiki/Cambridge_Analytica#2016_presidential_election). ***Page Public Content Access*** became only available to approved to only verified "businesses" for reasons that make no sense as the name suggests: access to page content that is already public. "Individual" entities are only allowed the most limited accesses as of early-2019 and have no ***Page Public Content Access***. When contacted, Facebook support equivalented sole propriertor entity to an "individual" entity. Subsequently, the new mid-2018 "App Review" process was never completed and Graph API access was blocked. 

I am releasing this code under MIT License in hope that the code helps you with your projects. Also because Fly Space-A is not running due to the above issue.

How to use
-------------

1. Install Go

2. `go get https://github.com/ansonl/flyspacea-backend`

3. Paste your Facebook Graph API Access Token into `constants.go`.

4. Set $DATABASE_URL to your PostgreSQL database URL.

3. `go install spacea`

4. `spacea -procMode=all`

Debug Mode Notes
-------------
All the constants mentioned below are located in `constants.go`.

- To generate updated timezones for a set of terminals with latitude and longitude inputed into the terminal JSON file (set at `TERMINAL_FILE`, set `DEBUG_EXPORT_TERMINAL_TZ` to *true*. 

- To run photo processing on a single terminal's photos, set `DEBUG_TERMINAL_SINGLE_FILE` to *true* and place the individual terminal JSON data into the filename set at `TERMINAL_SINGLE_FILE`. This will download the terminal's photo from the associated Facebook page specified by Facebook ID and process the photos into flight data. 

- To run photo processing on local images, set `DEBUG_MANUAL_IMAGE_FILE_TARGET` to *true*. This will make Fly Space-A process images in the directory set as `DEBUG_MANUAL_IMAGE_FILE_TARGET_TRAINING_DIRECTORY` with the extension set as `DEBUG_MANUAL_FILENAME`. 

*Recommend first skimming [technical implementation](https://docs.google.com/presentation/d/1cnS_nTL6xhL5PEHFro7jvDSuHAccr8eSBFV26KIfrzE/edit?usp=sharing) slides also available in `assets` folder for an overview of the photo processing steps. More information on debug modes can be obtained by searching for occurences of the debug constants in the entire project directory to find instances of debug constant usage. *

Credits
-------------

[jbowtie](https://github.com/jbowtie) fork of [gokogiri](https://github.com/jbowtie/gokogiri)

[latlng](github.com/bradfitz/latlong) by bradfitz

[pq](github.com/lib/pq) - Golang PostgreSQL driver

[Fuzzy](https://github.com/sajari/fuzzy) by Sajari

[goprocinfo](https://github.com/c9s/goprocinfo) by c9s

[ImageMagick](https://github.com/ImageMagick/ImageMagick) by [ImageMagick Studios LLC](https://imagemagick.org/)
