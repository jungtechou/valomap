# Development stage for local development and setup
FROM node:18-alpine AS development

WORKDIR /app

# Copy package.json and install dependencies
COPY package*.json ./

# We'll mount the source code as a volume in docker-compose

# Test stage for running tests
FROM node:18-alpine AS test

WORKDIR /app

# Copy package.json and install dependencies
COPY package*.json ./
RUN npm install

# We'll mount the source code as a volume in docker-compose

# End-to-end test stage
FROM node:18-alpine AS e2e

WORKDIR /app

# Install Cypress dependencies
RUN apk add --no-cache xvfb ttf-freefont chromium

# Copy package.json and install dependencies including Cypress
COPY package*.json ./
RUN npm install

# We'll mount the source code as a volume in docker-compose

# Production build stage
FROM node:18-alpine AS build

WORKDIR /app

# Copy package.json and install dependencies
COPY package*.json ./
RUN npm install

# Copy the rest of the application
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built app to nginx
COPY --from=build /app/build /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
