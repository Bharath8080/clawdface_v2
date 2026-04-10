FROM 605982976738.dkr.ecr.us-east-2.amazonaws.com/node:24.7-alpine
WORKDIR src/app
COPY package.json ./
COPY package-lock.json ./

RUN npm install --legacy-peer-deps

ENV NODE_ENV=production
ENV PORT=3079
COPY . .


# Increase Node.js memory limit for build
ENV NODE_OPTIONS="--max-old-space-size=4096"

RUN npm run build
EXPOSE 3079

CMD ["npm", "run", "start"]