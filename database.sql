DROP TABLE IF EXISTS flights;
CREATE TABLE HR72_flights (
Origin VARCHAR(100),
Destination VARCHAR(100),
RollCall TIMESTAMP,
UnknownRollCallDate BOOLEAN,
SeatCount INT,
SeatType VARCHAR(3), 
Cancelled BOOLEAN,
PhotoSource VARCHAR(2048),
SourceDate TIMESTAMP,
CONSTRAINT flights_pk PRIMARY KEY (Origin, Destination, RollCall, PhotoSource),
CONSTRAINT flights_origin_fk FOREIGN KEY (Origin) REFERENCES Locations(Title),
CONSTRAINT flights_dest_fk FOREIGN KEY (Destination) REFERENCES Locations(Title));

DROP TABLE IF EXISTS Locations;
CREATE TABLE Locations (
 Title VARCHAR(100),
 Phone VARCHAR(50),
 Email VARCHAR(255),
 GeneralInfo VARCHAR(2048),
 FBId VARCHAR(255),
 URL VARCHAR(2048),
 CONSTRAINT locations_pk PRIMARY KEY (Title));

#Insert new locations
INSERT INTO locations (Title, URL) 
    VALUES ($1, $2);

    INSERT INTO locations (Title, Phone, Email, GeneralInfo, FBId, URL) 
	    	VALUES ("test", "test", "test", "test", "test", "test") 
	    	ON CONFLICT (Title) DO UPDATE SET
	    	Phone = EXCLUDED.Phone,
	    	Email = EXCLUDED.Email,
	    	GeneralInfo = EXCLUDED.GeneralInfo,
	    	FBId = EXCLUDED.FBId;

#Insert new flight
INSERT INTO flights (Origin, Destination, RollCall, SeatCount, SeatType, Cancelled, PhotoSource, SourceDate) 
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

SELECT Origin, Destination, RollCall, SeatCount, SeatType, Cancelled 
 FROM flights 
 WHERE Origin=$1 AND RollCall <= current_timestamp + INTERVAL '14 day' AND RollCall > current_timestamp - INTERVAL '14 day';

SELECT Origin, Destination, RollCall, SeatCount, SeatType, Cancelled 
 FROM flights 
 WHERE Origin=$1 AND RollCall <= to_timestamp('2016-04-18 17:50:00', 'YY-MM-DD HH24:MI:SS') + INTERVAL '14 day' AND RollCall > to_timestamp('2016-04-18 17:50:00', 'YY-MM-DD HH24:MI:SS') - INTERVAL '14 day';

#delete in current day
DELETE FROM %v 
 WHERE Origin=$1 AND RollCall >= now() AND RollCall <= SELECT CURRENT_DATE + INTERVAL '1 DAY';

#delete in future day
DELETE FROM %v 
 WHERE Origin=$1 AND RollCall >= $2 AND RollCall <= $2 + INTERVAL '1 DAY';


 RollCall > to_timestamp('2016-04-18 17:50:00', 'YYYY-MM-DD') - INTERVAL '14 day';

 #delete in future day
DELETE FROM %v 
 WHERE Origin=$1 AND RollCall >= $2 AND RollCall < $3;


#Insert new flight
INSERT INTO %v (Origin, Destination, RollCall, SeatCount, SeatType, Cancelled, PhotoSource) 
    VALUES ($1, $2, $3, $4, $5, $6, $7);

#select flights from origin
SELECT Origin, Destination, RollCall, SeatCount, SeatType, Cancelled, PhotoSource
FROM %v
WHERE Origin=$1 AND RollCall >= $2 AND RollCall < $3;

#select flights to destination
SELECT Origin, Destination, RollCall, SeatCount, SeatType, Cancelled, PhotoSource
FROM %v
WHERE Destination=$1 AND RollCall >= $2 AND RollCall < $3;

#select flights from origin to destination
SELECT Origin, Destination, RollCall, SeatCount, SeatType, Cancelled, PhotoSource
FROM %v
WHERE Origin=$1 AND Destination=$2 AND RollCall >= $3 AND RollCall < $4;

#select flights from time start to end
SELECT Origin, Destination, RollCall, SeatCount, SeatType, Cancelled, PhotoSource
FROM %v
WHERE RollCall >= $1 AND RollCall < $2;