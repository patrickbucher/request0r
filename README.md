# request0r

Request URL `https://go.dev/` with `3` workers, performing `5` requests each (15 requests):

    $ go run main.go -w 3 -r 5 https://go.dev/
    240.053966ms

The duration 240.053966ms is the mean request time.