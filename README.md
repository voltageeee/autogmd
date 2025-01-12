# autogmd
не забудь создать .env файл если захочешь потестить:
```
STEAM_API_KEY=твойстимвебапиключь
DB_CONNECTION_STRING=user:password@tcp(ip:port)/database
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
