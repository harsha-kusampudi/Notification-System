# Design

- After receiving the input (either through web or cli), the notification along with the scheduled time is stored in the dynamoDB.
    - If the schedule time is some time today, then the notification is straightaway pushed to the redis.
- A go routine `CheckForDueNotifications` continuous runs to check for any due notifications in the redis and pushes to a channel if any.
- There is a worker pool, where each worker runs, reads from the channel and processes the notifications (processing in my implementation corresponds to just printing it out to stdout).
- There is a go routine `D2R_everyday`, which runs at midnight (00:00:00 IST), and queries the dynamoDB, if any entries are present for that day, they are pushed to redis.
- The reason to do this is to not have everything in the redis, because it is more resource expensive. However there is a tradeoff with the routine that runs everyday.

## DynamoDB Design:
- The Partition key is kept to be event_type, i.e., notification here, though it is unnecessary in our case, but if there's any future extension for other types of events, this helps.
- The Sort Key is the notification_time#notification_ID, this is done to effectively query all the notifications within an interval and to maintain uniqueness of the primary key, notification_ID is also appended.

## Redis Design:
- Used a set to store the notification with the score as the time and member as the notification id
- Used a hash to store the notification with the notification id as the key and the message and time as the value

## Improve?
- Ideally, the `CheckForDueNotifications` go routine can be implented as a cron, which runs every second, as here the smallest unit of notification time is second only (Can't schedule notifications in milliseconds).
- The `D2R_everyday` can also be a cron which runs everyday at midnight.

## Steps to run

### Necessary installations
- Install redis
- Start redis using `redis-server`
- Install dynamodb-local
- Start dynamodb local using `java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -sharedDb`

### To run with web interface

##### Starting the frontend interface
- Install necessary packages using `npm install`
- Start the frontend server using `npm start`

##### Starting the backend server
- `go run cmd/main/*.go -mode=web`

### To run with CLI

##### Starting the server
- `go run cmd/main/*.go -mode=cli`
- Select 1 to continue entering the notification message
    - Enter the message (in one line)
    - Enter the notification time in YYYY:MM:DD HH:MM:SS
- Select 2 to exit the server

