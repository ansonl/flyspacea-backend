DROP TABLE IF EXISTS flights;
CREATE TABLE flights (
 Origin VARCHAR(50),
 Destination VARCHAR(50),
 RollCall TIMESTAMP,
 SeatCount INT,
 SeatType VARCHAR(3), 
 Cancelled BOOLEAN,
 CONSTRAINT flights_pk PRIMARY KEY (Origin, Destination, RollCall));

#Insert new pickup
INSERT INTO flights (Origin, Destination, RollCall, SeatCount, SeatType, Cancelled) 
    VALUES ($1, $2, $3, $4, $5, $6);

SELECT Origin, Destination, RollCall, SeatCount, SeatType, Cancelled 
 FROM flights 
 WHERE Origin=$1 AND RollCall <= current_timestamp AND RollCall > current_timestamp - INTERVAL '14 day';