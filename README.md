# MX HTTP Proxy
Приложение запускается как сервис и позволяет выполнять через [REST API](https://en.wikipedia.org/wiki/Representational_state_transfer) команды на удаленном сервере [Zultys MX](https://www.zultys.com/zultys-cloud-services/). Для отслеживания событий сервера Zultys MX используется [Server-Sent Events API](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events).

## Описание API
Сервер включает в себя [описание поддерживаемого API](docs/openapi.yaml) в виде документа в формате [OpenAPI 3.0](https://github.com/OAI/OpenAPI-Specification). Так же поддерживается визуальное представление данного документа при обращении к корню сервиса.

<details>
<summary><strong>Пример использования API</strong></summary>

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
```
```http
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
</details>

## Параметры запуска сервера
Все настройки сервиса осуществляются через параметры для запуска.

<details>
<summary><code>-port &lt;port></code></summary>

Задает имя хоста (опционально) и порт, на котором будет отвечать HTTP-сервер.

Так же может быть задано через переменную окружения `PORT`.
</details>
<details>
<summary><code>-mx &lt;mxhost></code></summary>

Задает адрес сервера Zultys MX. Если порт не указан, то по умолчанию используется порт `7778`.

Так же может быть задано через переменную окружения `MX`:

    $ export MX=mxhost.connector73.net

**Внимание:** незащищенное соединение с сервером Zultys MX не поддерживается!
</details>
<details>
<summary><code>-letsencrypt &lt;hostname></code></summary>

Задает имя хоста для поддержки HTTPS. Сертификат будет автоматически получени и при необходимости обновлен с помощью сервиса [Let's Encrypt](https://letsencrypt.org). Можно указать сразу несколько имен хостов, разделив их запятыми.

**Внимание:** в данном случае задание порта для HTTP-сервера будет проигнорировано, т.к. для поддержки нормальной работы получения и обновления сертификата необходимо, чтобы сервер был настроен на 80 и 443 порты.

 Так же может быть задано через переменную окружения `LETSENCRYPT_HOST`.
</details>
<details>
<summary><code>-log &lt;params></code></summary>
Задает настройки для вывода лога работы сервиса.

Для вывода лога используется `stderr`. Если необходимо переопределить вывод лога в файл, то можно воспользоваться следующим методом:

    $ ./mx-http-proxy 2>mx-http-proxy.log

Вы можете задать уровень сообщений для вывода в лог:

- `all` - выводить все записи лога
- `trace` - выводить все записи лога, начиная от команд и событий от сервера MX
- `debug` - выводить все записи лога, начиная с отладочных выводов, но исключая вывод команд и событий сервера MX
- `info` - выводить все записи лога, начиная от информационных, но исключая отладочные данные
- `warn` - не выводить информационных сообщений лога, а только об ошибках и предупреждениях
- `error` - выводить только сообщения об ошибках
- `none` - вообще отключить вывод лога

Так же можно задать формат вывода лога:

- `json` - использовать формат JSON для вывода лога
- `console` - использовать консольный формат для вывода лога
- `color` - использовать консольный формат с цветовым выделением для вывода лога
- `develop` - аналогичен формату `color`, но значения атрибутов лога выводит с новой строки

По умолчанию используется консольный формат и выводятся все информационные сообщения, предупреждения, а так же сообщения об ошибках. Отладочные сообщения и команды с событиями сервера MX в лог не выводятся, если это явно не задано.

Можно задать сразу несколько значений параметра лога, указав их через запятую или двоеточие:

    ./mx-http-proxy -log dev,all

Настройки лога по умолчанию можно изменить для всех приложений, задав их в виде переменной окружения `LOG`:

    $ export LOG=COLOR

Или только для однократного запуска приложения:

    $ LOG=DEV ./mx-http-proxy -mx localhost

Вывод лога в формате [JSON](https://www.json.org) позволяет легче разбирать его программным образом:

    $ ./mx-http-proxy -log json,all
    {"ts":1533494002,"lvl":0,"msg":"service","name":"MX-HTTP-Proxy","version":"dev"}
    {"ts":1533494002,"lvl":0,"log":"http","msg":"server","listen":"localhost:8000","tls":false,"url":"http://localhost:8000/"}

Сообщения, относящиеся к командам сервера MX используют тип лога - `mx`, а для информации об обработке HTTP-запросов используется - `http`. Все остальные выводы обычно не используют префикс:

    21:52:50.512539 INFO  service name=MX-HTTP-Proxy version=dev date=2018-08-05 build=063525b
    21:52:50.512949 INFO  [http]: server listen=localhost:8000 tls=false url=http://localhost:8000/
    21:53:00.036338 TRACE [mx]: dmtest3: <- 0001 <loginRequest type="User" platform="iPhone" version="7.0" loginCapab="Audio" mediaCapab="Voicemail|CallRec"><userName>dmtest3</userName><pwd>nnke/C/yi/f...U5OVTqg5joXHc=&#xA;</pwd></loginRequest>
    21:53:00.070054 TRACE [mx]: dmtest3: -> 0001 <loginResponce Code="0" sn="631HC" apiversion="11" ext="273" userId="43892780322813134" softPhonePwd="yTEuJ15RheF2...BogZXzp27fAqc334X"  proto="TLS" mxport="5061" clientport="1234" >Login OK</loginResponce>
    21:53:00.070779 DEBUG store connection login=dmtest3 token=m_r3PY2i1jkxNo6W
    21:53:00.071176 INFO  [http]: POST /login code=200 user=dmtest3 size=180 duration=144.83142ms gzip=false
    21:53:01.821825 TRACE [mx]: dmtest3: <presence from="0" status="Available" mxStatus=""></presence>
    21:53:01.822540 DEBUG sse user=dmtest3 event=presence data={"presence":"Available"}
</details>

## Локальные сертификаты

Если в каталоге `./certs/` найдены пары сертификатов (`cert.key` и `cert.crt`), то они будут автоматически загружены и использованы веб-сервером для поддержки HTTPS. Таких пар сертификатов может быть несколько.

Это могут быть и самоподписанные сертификаты. Например, для создания сертификата для `localhost` можно воспользоваться следующей командой:

    openssl req -x509 -out localhost.crt -keyout localhost.key \
	    -newkey rsa:2048 -nodes -sha256 \
	    -subj '/CN=localhost' -extensions EXT -config <( \
	    printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

**Внимание:** в этом случае необходимо явно указывать, что используется протокол `https`:

    $ curl https://localhost:8000/ -k

## Поддержка Docker

Данный сервис доступен в виде образов Docker:

    $ docker pull mdigger/mx-http-proxy
    $ docker run --rm \
        -p 8000:8000 \
        -e MX=631hc.connector73.net \
        mdigger/mx-http-proxy -log all,color

Кроме описанных выше параметров, сборка в Docker может потребовать подключения хранилища для кеша сертификатов `/letsEncrypt.cache` и/или хранилища сертификатов `/certs`. В последнем случае сертификаты из каталога будут автоматически загружены и сервер будет отвечать исключительно по протоколу `https://`.

## Мониторинг

По адресу `/debug/vars` доступны описание некоторых метрик сериса в формате JSON, которые могут быть использованы для мониторинга сервиса.
