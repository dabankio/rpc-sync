####  
```
protoc ./dbp/dbp.proto --python_out=./
protoc ./dbp/lws.proto --python_out=./
protoc ./dbp/sn.proto --python_out=./
```

#### 
``` sh
sudo docker ps -a
sudo docker exec -it 3a7c83e6b635 /bin/bash
pip install pymysql
pip install protobuf
```
