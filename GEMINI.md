# Project Abstract: AnyAdmin

## Overview
[Provide a 1-2 paragraph summary of what AnyAdmin is. What problem does it solve? What is its main purpose?]

## Project Structure
/backend: Go-based API service (Gin framework) handling orchestration and node management.
/frontend: Node.js/Express application with Pug templates for the management UI.


## Coding Guide
* **Unit Test**: make all your unit test under @tests\ directory, Write unit tests to verify any update systematically by accessing the frontend app and backend services code changes. Make sure you pass all the unit tests. if failed any unit test, you should fix it or retry another approach. You should use the remote host on this ip 172.25.208.100 with port 8082 for agent interaction 
* **logging**: all the loggging should be put under /logs folder
* **ssh key**: scan this directory for ssh key @backend\keys
* **Complie**: - You have to recompile the agent to @backend/dist/anyadmin-agent, and then redeploy to the target machine if any changes made on agent related code. restart the frontend and backend processes if needed.
* **Code Execution**: 
- Launch the apps / processes with seperate process and redirect the output to logs under /logs folder for debugging, such that you avoids single-threaded debug mode and prevents getting stuck at waiting feedbacks.
- Avoid using `taskkill` broadly, as it may terminate the CLI itself. Instead, use  powershell scripts under @scripts folder to terminate proccess on port 3000 and 8080 or redeploym agents


## Environment
* **Execution Environment**: This is powershell, windows 11.
* **Remote Agent Environment**: The remote host is up and running at 172.20.0.10 on port 22 at path /home/anyadmin/bin, and you have to run ssh with passwordless id_rsa key at @backend\keys with root user to remote host. 