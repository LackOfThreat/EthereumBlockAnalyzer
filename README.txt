This application requires database. To create a docker container with postgresql, 
please proceed to the folder where you saved this repository and run: 

"docker-compose up -d db"

After that you can run the server (it will run on localhost) and paste the http endpoint /api/block/NEEDED_ETH_BLOCK_NUMBER/total

To check the database and data stored in it you can do next:
1. Open terminal.
2. Type "docker ps". You will see the id of the container that is running now.
3. Connect to it "docker exec -it id_of_container psql -U postgres"
4. Connect to database "\c EthereumData"
5. Name of the tabke is "blocks", so to see all data from table just type "SELECT * FROM blocks;"

Good luck!