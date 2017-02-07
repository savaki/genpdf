FROM ubuntu:16.04
MAINTAINER Altair Six

RUN apt-get update && apt-get install -y nodejs wget unzip telnet awscli npm phantomjs

RUN ln -s /usr/bin/nodejs /usr/bin/node

RUN npm install -g html-pdf

ADD html-pdf.js /usr/local/lib/node_modules/html-pdf/bin/html-pdf.js
RUN ln -s /usr/local/lib/node_modules/html-pdf/bin/html-pdf.js /usr/local/bin


