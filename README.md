# MX HTTP Proxy
Приложение запускается как сервис и позволяет выполнять через [REST API](https://en.wikipedia.org/wiki/Representational_state_transfer) команды на удаленном сервере [Zultys MX](https://www.zultys.com/zultys-cloud-services/). Для отслеживания событий сервера Zultys MX используется [Server-Sent Events API](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events).

## Описание API
Сервер включает в себя описание поддерживаемого API в виде документа в формате [OpenAPI 3.0](https://github.com/OAI/OpenAPI-Specification) [`openapi.yaml`](www/openapi.yaml). Так же поддерживается и визуальное представление данного документа при обращении к корню сервиса.

Пример использования API:

```http
POST /login HTTP/1.1
Host: localhost:8000
Content-Type: application/json; charset=utf-8
Content-Length: 144

{
    "login":"login",
    "password":"password",
    "type":"User",
    "platform":"iPhone",
    "version":"7.0",
    "loginCapab":"Audio",
    "mediaCapab":"Voicemail|CallRec"
}

HTTP/1.1 200 OK
Server: MX-HTTP-Proxy/0.1.3 (2ce6c32)
X-API-Version: 1.0.2
Access-Control-Allow-Origin: *
Content-Type: application/json; charset=utf-8
Content-Length: 180

{
    "token": "81snQUFPMDs7GEye",
    "user": "43892780322813134",
    "device": "273",
    "softPhonePwd": "nfsi8ohraw2ReJtjCuE7f3KyTWc2doUi",
    "api": 11,
    "mx": "631HC"
}
```

## Параметы запуска сервера
Все настройки сервиса осуществляются через параметры для запуска.

### `-http <hostname>`
Задает имя хоста и порт, на котором будет отвечать HTTP-сервер. Если указано доменное имя хоста (не `localhost`, и не IP-адрес), то настраивается автоматическое получение сертификата для данного домена через сервис [Let's Encrypt](https://letsencrypt.org).

**Примеры**:
- `localhost:8000` - только локальный доступ
- `:8000` - с любого хоста
- `myserver.examle.com:8000` - только для указанного имени домена
- `myserver.examle.com` - автоматическая поддержка HTTPS
- `myserver.examle.com:443` - аналогично предыдущему варианту

Вместо автоматического получения сертификатов вы можете их явно указать в параметрах запуска.

### `-certs public.crt,private.key`
Позволяет указать локальные файлы с сертификатами для поддержки HTTPS. В этом случае протокол HTTPS может использоваться для любого имени хоста и порта. Автоматическое получение сертификата через Let's Encrypt при этом отключается.

Для тестирования можно использовать самоподписанные сертификаты.

### `-mx mxhost.connector73.net`
Задает адрес для сервера Zultys MX. Если порт не указан, то по умолчанию используется порт `7778`.

**Внимание:** незащищенное соединение с сервером Zultys MX не поддерживается!

### `-log`

Задает настройки для вывода лога работы сервиса.

Поддерживаются следующие значения:

- `ALL` - выводить все записи лога (синонимы: `A`, `*`)
- `TRACE` - выводить все записи лога, начиная от команд и событий от сервера MX (синонимы: `TRC`, `T`)
- `DEBUG` - выводить все записи лога, начиная с отладочных выводов, но исключая вывод команд и событий сервера MX (синонимы: `DBG`, `D`)
- `INFO` - выводить все записи лога, начиная от информационных, но исключая отладочные данные (синонимы: `INF`, `I`)
- `WARNING` - не выводить информационных сообщений лога, а только об ошибках и предупреждениях (синонимы: `WARN`, `WRN`, `W`)
- `ERROR` - выводить только сообщения об ошибках (синонимы: `ERR`, `E`)
- `NONE` - вообще отключить вывод лога (синонимы: `NO`, `N`, `OFF`)
- `JSON` - использовать формат JSON для вывода лога (синонимы: `JSN`, `J`)
- `CONSOLE` - использовать консольный формат для вывода лога (по умолчанию, синонимы: `STANDART`, `STD`, `S`)
- `COLOR` - использовать консольный формат с цветовым выделением для вывода лога (синонимы: `COLOR`, `COL`, `C`)
- `DEVELOPER` - аналогичен формату `COLOR`, но значения атрибутов лога выводит с новой строки (синонимы: `DEVELOPERS`, `DEVELOP`, `DEV`)

По умолчанию используется консольный формат и выводятся все информационные сообщения, предупреждения, а так же сообщения об ошибках. Отладочные сообщения и команды с событиями сервера MX в лог не выводятся, если это явно не задано.

Можно задать сразу несколько значений параметра лога, указав их через запятую или двоеточие:

    ./mx-http-proxy -log dev,all

Так же настройки лога по умолчанию можно изменить для всех приложений, задав их в виде переменной окужения `LOG`:

    $ export LOG=COLOR

Или только для однократного запуска приложения:

    $ LOG=DEV ./mx-http-proxy -mx localhost

Вывод лога в формате JSON позволяет легче разбирать его программным образом:

    $ ./mx-http-proxy -log json,all
    {"ts":1533494002,"lvl":0,"msg":"service","name":"MX-HTTP-Proxy","version":"dev"}
    {"ts":1533494002,"lvl":0,"log":"http","msg":"server","listen":"localhost:8000","tls":false,"url":"http://localhost:8000/"}

Сообщения, относящиеся к командам сервера MX используют тип лога - `mx`, а для информации об обработке HTTP-запросов используется - `http`. Все остальные выводы обычно не используют префикс:

    21:52:50.512539 INFO  service name=MX-HTTP-Proxy version=dev built=2018-08-05 commit=063525b
    21:52:50.512949 INFO  [http]: server listen=localhost:8000 tls=false url=http://localhost:8000/
    21:53:00.036338 TRACE [mx]: dmtest3: <- 0001 <loginRequest type="User" platform="iPhone" version="7.0" loginCapab="Audio" mediaCapab="Voicemail|CallRec"><userName>dmtest3</userName><pwd>nnke/C/yi/f...U5OVTqg5joXHc=&#xA;</pwd></loginRequest>
    21:53:00.070054 TRACE [mx]: dmtest3: -> 0001 <loginResponce Code="0" sn="631HC" apiversion="11" ext="273" userId="43892780322813134" softPhonePwd="yTEuJ15RheF2...BogZXzp27fAqc334X"  proto="TLS" mxport="5061" clientport="1234" >Login OK</loginResponce>
    21:53:00.070779 DEBUG store connection login=dmtest3 token=m_r3PY2i1jkxNo6W
    21:53:00.071176 INFO  [http]: POST /login code=200 user=dmtest3 size=180 duration=144.83142ms gzip=false
    21:53:01.821825 TRACE [mx]: dmtest3: <presence from="0" status="Available" mxStatus=""></presence>
    21:53:01.822540 DEBUG sse user=dmtest3 event=presence data={"presence":"Available"}