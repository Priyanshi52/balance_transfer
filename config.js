var util = require('util');
var path = require('path');
var hfc = require('fabric-client');

//var Client = new hfc();
//Client.setConfigSetting('request-timeout', 120000);

var file = 'network-config%s.json';

var env = process.env.TARGET_NETWORK;
if (env)
	file = util.format(file, '-' + env);
else
	file = util.format(file, '');

hfc.addConfigFile(path.join(__dirname, 'app', file));
hfc.addConfigFile(path.join(__dirname, 'config.json'));