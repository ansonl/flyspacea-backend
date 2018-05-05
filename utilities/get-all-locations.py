import requests
import re
import json

usaLink = "http://spacea.net/usa-locations"
f = requests.get(usaLink)
usaStrip = f.text.replace('\n', '').replace('\r', '')

europeLink = "http://spacea.net/europe-locations"
f = requests.get(europeLink)
europeStrip = f.text.replace('\n', '').replace('\r', '')

pacificLink = "http://spacea.net/pacific-locations"
f = requests.get(pacificLink)
pacificStrip = f.text.replace('\n', '').replace('\r', '')

otherLink = "http://spacea.net/other-locations"
f = requests.get(otherLink)
otherStrip = f.text.replace('\n', '').replace('\r', '')

reProvince = re.compile('<td class="views-field views-field-province" >[ ]*([a-zA-Z ]*)[ ]*<\/td>')

reCountry = re.compile('<td class="views-field views-field-country" >[ ]*([a-zA-Z ]*)[ ]*<\/td>')

reLocation = re.compile('<td class="views-field views-field-title" >[^<>]*<a[^<>]*>([.,a-zA-Z()&#;\'\/\- 0-9]*)[ ]*<\/a>')

#build combined location string list
combinedLocations = []

matchProvince = reProvince.findall(usaStrip)
matchLocation = reLocation.findall(usaStrip)
for index, item in enumerate(matchLocation):
  combinedString = matchLocation[index]
  #print(index)
  #print(matchLocation[index])
  #print(matchProvince[index])
  combinedString += ", " +  matchProvince[index]
  combinedLocations.append(combinedString.strip())

matchCountry = reCountry.findall(europeStrip)
matchLocation = reLocation.findall(europeStrip)
for index, item in enumerate(matchLocation):
  combinedString = matchLocation[index]
  #print(index)
  #print(matchLocation[index])
  #print(matchProvince[index])
  combinedString += ", " +  matchCountry[index]
  combinedLocations.append(combinedString.strip())

matchCountry = reCountry.findall(pacificStrip)
matchLocation = reLocation.findall(pacificStrip)
for index, item in enumerate(matchLocation):
  combinedString = matchLocation[index]
  #print(index)
  #print(matchLocation[index])
  #print(matchProvince[index])
  combinedString += ", " +  matchCountry[index]
  combinedLocations.append(combinedString.strip())

matchCountry = reCountry.findall(otherStrip)
matchLocation = reLocation.findall(otherStrip)
for index, item in enumerate(matchLocation):
  combinedString = matchLocation[index]
  #print(index)
  #print(matchLocation[index])
  #print(matchProvince[index])
  combinedString += ", " +  matchCountry[index]
  combinedLocations.append(combinedString.strip())

#create object for JSON encoding
locations = []

for locationString in combinedLocations:
  locations.append({
    "title": locationString,
    "keywords": []
  });
  
  
#print(json.dumps(locations, indent="\t"))
print(json.dumps(locations, indent="\t"), file=open("usa-locations-spaceanet.json", "w"))