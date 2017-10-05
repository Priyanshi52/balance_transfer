FROM node:6.11.2

# Create app directory
run ["mkdir", "-p", "/usr/src/app"]
WORKDIR /usr/src/app
# WORKDIR /usr/src/app

# Install app dependencies
COPY package.json /usr/src/app/
# For npm@5 or later, copy package-lock.json as well
# COPY package.json package-lock.json ./

RUN npm install

# Bundle app source
COPY . /usr/src/app/

EXPOSE 4000
CMD [ "npm", "start" ]