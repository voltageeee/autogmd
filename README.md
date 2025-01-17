# autogmd
не забудь создать .env файл если захочешь потестить:
```
STEAM_API_KEY=твойстимвебапиключь
DB_CONNECTION_STRING=user:password@tcp(ip:port)/database
HOSTNAME=http://localhost:8080
```

запуск: 
```
go mod tidy
go run main.go
```

# /api/projects/get
возвращает проекты чубрика в джсон формате (array of Project object). метод get, единственный нужный хедер - токен сессии (для авторизации и получения проектов чубрика из базы данных) который автоматом отправляется бразуером.
пример респонса:
```
[
    {
        "ID": 9,
        "Name": "dasd",
        "IPAddr": "localhost",
        "Balance": 0,
        "Owner": "76561198969757101",
        "Secret": "82cQWBS-zYIe5ud0rR6M9kOsHu6cNGvSOQZU1Wsv8MM="
    }
]
```

# /api/projects/create
создает проект. метод post, нужна форма с ключом "projectname", значение которого - имя проекта. единственный хедер - токен сессии. респонса нема, но должен вернуть 200. иначе беда!! имя должно
быть уникальным.
пример запроса:
![image](https://github.com/user-attachments/assets/2b1ac235-5c33-4c30-a9e1-52f8390e3be0)

# /api/projects/delete
удаляет проект, метод post, нужна форма с ключем "projectname", значение - имя проекта. хедер как обычно - токен сессии. респонса тож нема, но должен вернуть 200. пример запроса:
![image](https://github.com/user-attachments/assets/e9be893b-ffed-44a5-b025-1757eb745f03)

# /api/projects/edit
меняет имя проекта по сути, метод post, нужна форма с ключем "projectid" (там собсна id проекта) и "newname" с новым именем.
пример:
![image](https://github.com/user-attachments/assets/d67c41f9-4208-426a-9322-52d866fbd919)


# /api/items/create
создает итем. метод post, нужна форма с ключами как на скрине. единственный хедер - токен сессии. респонса нема, должен вернуть 200. пример запроса: previousprice может быть 0, но тогда в магазе нужно будет previousprice ваще не отображать. итемы могут быть неуникальными) ещё надо category, если null - будет "Разное"
![image](https://github.com/user-attachments/assets/3e6a32b6-bfc5-4b45-b54d-488b143299ea)

# /api/items/get
получает итемы по проекту. !!!! паблик ендпойнт (не требует авторизации), метод гет. нужен ключ с айдишником проекта. пример респонса: (так-же есть category вида string)
![image](https://github.com/user-attachments/assets/dd62dad7-c40d-45a4-8f9b-8302e52592dc)

# /api/items/edit
изменяет итем. можно изменить картинку, имя, описание, цену, старую цену, категорию. пример:
![image](https://github.com/user-attachments/assets/142f58c4-6d5a-4473-80e1-416bda6c0a7e)

# /api/items/delete
удаляет итем, нужна форма с ключем itemid
