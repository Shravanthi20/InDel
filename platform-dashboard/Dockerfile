FROM node:18-alpine

WORKDIR /app

ARG VITE_PLATFORM_API_URL=http://localhost:8004
ENV VITE_PLATFORM_API_URL=${VITE_PLATFORM_API_URL}

COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

FROM node:18-alpine
RUN npm install -g serve

WORKDIR /app
COPY --from=0 /app/dist ./dist

EXPOSE 5174
CMD ["serve", "-s", "dist", "-l", "5174"]
