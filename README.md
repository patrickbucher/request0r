# request0r

Request URL `https://go.dev/` with `3` workers, performing `5` requests each (15 requests):

    $ go run main.go -w 3 -r 5 https://go.dev/
          mean time   requests     passed     failed
       820.648671ms         15         15          0
