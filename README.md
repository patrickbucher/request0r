# request0r

Requests an URL concurrently and reports statistics.

## Usage

Request URL `https://go.dev/` with `3` workers, performing `5` requests each (15 requests):

    $ go run main.go -w 3 -r 5 https://go.dev/
    Requests:
              Total          Passed          Failed            Mean
                 15              15               0    220.777262ms
    Percentiles:
                 0%             25%             50%             75%            100% 
        144.93501ms    166.733358ms    168.090036ms    205.507289ms    439.357751ms
