DROP TABLE IF EXISTS flights;
CREATE TABLE HR72_flights (
 Origin VARCHAR(50),
 Destination VARCHAR(50),
 RollCall TIMESTAMP,
 SeatCount INT,
 SeatType VARCHAR(3), 
 Cancelled BOOLEAN,
 PhotoSource VARCHAR(2048),
 CONSTRAINT flights_pk PRIMARY KEY (Origin, Destination, RollCall));

DROP TABLE IF EXISTS locations;
CREATE TABLE locations (
 Title VARCHAR(50),
 URL VARCHAR(2048),
 CONSTRAINT locations_pk PRIMARY KEY (Title));

#Insert new locations
INSERT INTO locations (Title, URL) 
    VALUES ($1, $2);

#Insert new flight
INSERT INTO flights (Origin, Destination, RollCall, SeatCount, SeatType, Cancelled, PhotoSource) 
    VALUES ($1, $2, $3, $4, $5, $6, $7);

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
 WHERE Origin=$1 AND RollCall >= $2 AND RollCall < $2;


#Insert new flight
INSERT INTO %v (Origin, Destination, RollCall, SeatCount, SeatType, Cancelled, PhotoSource) 
    VALUES ($1, $2, $3, $4, $5, $6, $7);