# funda-checker# funda-checker

example for `docker-compose`
```
  funda-checker:
    container_name: funda-checker
    restart: always
    image: vladikamira/funda-checker:v0.0.1
    command:
      - '-scrapeDelayMilliseconds=500'
      - '-fundaSearchUrl=https://www.funda.nl/koop/amsterdam/300000-440000/70+woonopp/2+slaapkamers/'
      - '-fileNameToStoreResutls=/funda-checker/results.gob'
      - '-postCodes=1185,1186'
      - '-telegramToken=API_TOKEN_SECRET'
      - '-telegranChatId=AWESOME_CHAT_ID'
    volumes:
      - ./funda-checker:/funda-checker
    environment:
      - GOMAXPROCS=1
```
