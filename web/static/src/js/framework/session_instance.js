hexya.define('web.session', function (require) {
    var Session = require('web.Session');
    var modules = hexya._modules;
    return new Session(undefined, undefined, {modules:modules, use_cors: false});
});
