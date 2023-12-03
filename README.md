# request0r

Requests an URL concurrently and reports statistics.

## Usage

Request URL `https://go.dev/` with `3` workers, performing `5` requests each (15 requests):

    $ go run main.go -w 5 -r 5 https://go.dev/
    Requests:
              Total          Passed          Failed            Mean
                 25              25               0    311.920786ms
    Percentiles:
                 0%             25%             50%             75%            100%
       156.408004ms    162.548377ms    168.005046ms    177.479443ms    180.391166ms
