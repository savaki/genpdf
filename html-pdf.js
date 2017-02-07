#!/usr/bin/env node

var fs = require('fs');
var pdf = require('../');
var path = require('path');

var args = process.argv.slice(2);

if (args.length >= 2) {
  htmlpdf(args[0], args[1])
} else {
  help()
}

function help() {
  var help = [
    'Usage: html-pdf <source> <destination>',
    'e.g.: html-pdf source.html destination.pdf'
  ].join('\n');

  console.log(help)
}

function htmlpdf(source, destination) {
  var html = fs.readFileSync(source, 'utf8');
  var options = {
    base: 'file://' + path.resolve(source),
    border: {
      top: ".5in",            // default is 0, units: mm, cm, in, px
      bottom: ".5in",
      right: ".25in",
      left: ".25in"
    }
  };
  pdf.create(html, options).toFile(destination, function (err, res) {
    if (err) throw err
  })
}
