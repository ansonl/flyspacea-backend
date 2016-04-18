DROP TABLE IF EXISTS flights;
CREATE TABLE flights (
 origin CHAR(50),
 destination CHAR(50),
 departTimestamp TIMESTAMP,
 seats INT,
 seatRelease CHAR(50), 
 CONSTRAINT flights_pk PRIMARY KEY (origin, destination, departTimestamp));

#Insert new pickup
INSERT INTO flights (origin, destination, departTimestamp)
  VALUES ('a', 'b', '2002-10-02T10:00:00-05:00');