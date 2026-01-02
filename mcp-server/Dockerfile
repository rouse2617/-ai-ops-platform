FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY tsconfig.json ./
COPY src ./src

RUN npm run build

EXPOSE 3001

ENV NODE_ENV=production
ENV PORT=3001
ENV BACKEND_URL=http://backend:8080

CMD ["node", "dist/index.js"]
