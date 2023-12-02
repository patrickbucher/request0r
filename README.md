# request0r

Request URL `https://go.dev/` with `3` workers, performing `5` requests each (15 requests):

    $ go run main.go -w 3 -r 5 https://go.dev/
            mean          25%          50%          75%   requests     passed     failed
     523.07818ms 503.101025ms 515.632276ms 515.632276ms         15         15          0
