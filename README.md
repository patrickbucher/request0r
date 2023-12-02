# request0r

Requests an URL concurrently and reports statistics.

## Usage

Request URL `https://go.dev/` with `3` workers, performing `5` requests each (15 requests):

    $ go run main.go -w 5 -r 20 https://go.dev/
    Requests:
              Total          Passed          Failed            Mean
                 15              15               0    918.857647ms
    Percentiles:
                 0%             25%             50%             75%            100% 
       883.396304ms    908.858273ms    915.573547ms    918.764777ms    922.190016ms
