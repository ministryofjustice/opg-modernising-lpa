var characters = '346789QWERTYUPADFGHJKLXCVBNM'

var generateRandom4Chars = function() {
    var result = [];

    for(var i = 0; i < 4; i++) {
        result.push(characters.charAt(Math.floor(Math.random() * characters.length)));
    }

    return result.join('');
}

//gross, but imposter doesn't support ES6 out of the box (also explains vars)
var uid = 'M-'+generateRandom4Chars()+'-'+generateRandom4Chars()+'-'+generateRandom4Chars()

respond()
    .withStatusCode(201)
    .withHeader('Content-Type', 'application/json')
    .withContent('{"uid":"'+uid+'"}');
