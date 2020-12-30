const express = require('express')
const app = express()
var bodyParser = require('body-parser')
app.use(bodyParser.json())
const port = 3011

const { createLogger, format, transports } = require('winston');
const { combine, label, printf } = format;
 
const myFormat = printf(({ level, message, label }) => {
  return `${message}`;
});
 
const logger = createLogger({
  format: combine(
    format.json(),
    myFormat
  ),
  transports: [new transports.File({ filename: 'combined.log' })]
});

/*
const logger = winston.createLogger({
  level: 'info',
  format: winston.format.json(),
  defaultMeta: { service: 'user-service' },
  transports: [
    //
    // - Write all logs with level `error` and below to `error.log`
    // - Write all logs with level `info` and below to `combined.log`
    //
    new winston.transports.File({ filename: 'error.log', level: 'error' }),
    new winston.transports.File({ filename: 'combined.log' }),
  ],
});
*/

app.get('/', (req, res) => {
    logger.info(`{"method" : "GET", "message" : "home"}`);
    res.send('Hello World!');
})

app.post('/usuario', (req, res) => {
    logger.info(`{"method" : "POST", "message" : "usuario"}`);
    res.send('usuario ok')
})

app.listen(port, () => {
  console.log(`Example app listening at http://localhost:${port}`)
})