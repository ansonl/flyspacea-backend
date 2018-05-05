import requests
import re
import json

link = "http://spacea.net/usa-locations"
f = requests.get(link)
strip = f.text.replace('\n', '').replace('\r', '')
#print(strip)


matchProvince = re.findall('<td class="views-field views-field-province" >[ ]*([a-zA-Z ]*)[ ]*<\/td>', strip)

matchLocation = re.findall('<td class="views-field views-field-title" >[^<>]*<a[^<>]*>([.,a-zA-Z()\- 0-9]*)<\/a>', strip)

print(matchProvince)

#build combined location string list
combinedLocations = []

for index, item in enumerate(matchLocation):
  combinedString = matchLocation[index]
  print(index)
  print(matchLocation[index])
  print(matchProvince[index])
  if len(matchProvince) > index:
    combinedString += ", " +  matchProvince[index]
  combinedLocations.append(combinedString)

#create object for JSON encoding
locations = []

for locationString in combinedLocations:
  locations.append({
    "title": locationString,
    "keywords": []
  });
  
  
print(json.dumps(locations, indent="\t"))
#print(json.dumps(locations, indent="\t"), file=open("usa-locations-spaceanet.json", "w"))