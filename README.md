# autogmd
не забудь создать .env файл если захочешь потестить:
```
STEAM_API_KEY=твойстимвебапиключь
DB_CONNECTION_STRING=user:password@tcp(ip:port)/database
HOSTNAME=http://localhost:8080
```

# /api/projects/get
возвращает проекты чубрика в джсон формате (array of Project object). метод get, единственный нужный хедер - токен сессии (для авторизации и получения проектов чубрика из базы данных) который автоматом отправляется бразуером.
пример респонса:
```
[{
    "ID": 2,
    "Name": "testproject",
    "IPAddr": "localhost",
    "Balance": 0,
    "Owner": "76561198969757101"
}, {
    "ID": 3,
    "Name": "testproject",
    "IPAddr": "localhost",
    "Balance": 0,
    "Owner": "76561198969757101"
}]
```

# /api/projects/create
создает проект. метод post, нужна форма с ключом "projectname", значение которого - имя проекта. единственный хедер - токен сессии. респонса нема, но должен вернуть 200. иначе беда!! имя должно
быть уникальным.
пример запроса:
![image](https://github.com/user-attachments/assets/2b1ac235-5c33-4c30-a9e1-52f8390e3be0)

# /api/projects/delete
удаляет проект, метод post, нужна форма с ключем "projectname", значение - имя проекта. хедер как обычно - токен сессии. респонса тож нема, но должен вернуть 200. пример запроса:
![image](https://github.com/user-attachments/assets/e9be893b-ffed-44a5-b025-1757eb745f03)

# /api/items/create
создает итем. метод post, нужна форма с ключами как на скрине. единственный хедер - токен сессии. респонса нема, должен вернуть 200. пример запроса: previousprice может быть 0, но тогда в магазе нужно будет previousprice ваще не отображать. итемы могут быть неуникальными, нам-то похуй че там сервер овнеры добавляют. вот прям поебать)
![image](https://github.com/user-attachments/assets/3e6a32b6-bfc5-4b45-b54d-488b143299ea)

# /api/items/get
получает итемы по проекту. !!!! паблик ендпойнт (не требует авторизации), метод гет. нужен ключ с айдишником проекта. пример респонса:
![image](https://github.com/user-attachments/assets/dd62dad7-c40d-45a4-8f9b-8302e52592dc)


