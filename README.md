# h5reduce

*h5reduce* is a very tiny and simple html5- and css reducer. it's main purpose
is to have something useable at hand without fetching a full blown jre to 
throw yui onto the problem (allthough yui yields better results).

## usage

flags:

    -css-nl=true: newlines after each rule
    -html-cdata=true: treat cdata as text-token (not as comment)
    -html-nl=true: place some additional newlines
    -i="": input file (default: stdin)
    -keep-excl-comments=true: don't strip away /*! or <!--! comments
    -o="": output file (default: stdout)
    -stats=false: print stats
    -strip-comments=true: strip away (most) comments

sample usage:

    $> h5reduce -i test.html -stats -o test-smaller.html
    0.76    24480   18532

## author

* mathias gumz <mg@2hoch5.com>

