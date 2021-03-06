# Deploying and Using Node Hibernator

## Deployment guidelines

* Node Hibernator must run on the same host as the linked Ethereum Client and Privacy Manager.

* Node Hibernator can be run as a shell process or a Docker container.  See (../README.md) for run commands.  This determines the available [process configuration](../config.md#process) options:
    * Shell process Node Hibernator: Can manage `shell` & `docker` Ethereum Client and Privacy Manager processes.
    * Docker container Node Hibernator: Can manage *only* `docker` Ethereum Client and Privacy Manager processes.

* Node Hibernator does not require a Privacy Manager.  All Privacy Manager related [config](../config.md) fields are optional.

* When using `shell` Ethereum Client and Privacy Manager:
    * If running multiple Ethereum Clients or Privacy Managers on the same host, ensure that Node Hibernator's configured start and stop scripts only impact a single node at a time to prevent unexpected behaviour
    * In start scripts redirect process output to a log file
    * In start scripts use `&` to run the process in the background 
    * See [samples/shell](samples/shell) for examples

## Adding Node Hibernator to an existing deployment

1. Construct the Node Hibernator config as required by the existing deployment 
1. Start Node Hibernator
1. If using Tessera: Update Tessera's server configs so that `serverAddress` is the corresponding Node Hibernator proxy address, and `bindingAddress` is the "internal" address that Node Hibernator will forward requests to. See [Tessera's Server Addresses docs](https://docs.tessera.consensys.net/en/latest/HowTo/Configure/TesseraAPI/#server-addresses) for more info. 
1. Update/inform clients to use the proxy addresses for all requests.  
   
**If clients continue to use the direct Ethereum Client and Privacy Manager API addresses instead of Node Hibernator's proxy addresses, Node Hibernator will be unable to accurately determine activity. This will likely lead to inconsistent behaviour.**

## Understanding Client Errors
The following table describes scenarios where user-submitted requests are expected to fail.  The Action describes the necessary steps to continue:

| Scenario  | Error | Action |
| --- | --- | --- |
| User sends request when Node Hibernator is hibernating the Ethereum Client and Privacy Manager | 500 (Internal Server Error) - `node is being shutdown, try after sometime` | Retry after some time. |  
| User sends request when Node Hibernator is starting the Ethereum Client and Privacy Manager | 500 (Internal Server Error) - `node is being started, try after sometime` | Retry after some time. |  
| User sends a private transaction request when at least one of the remote recipients is hibernated by Node Hibernator | 500 (Internal Server Error) - `Some participant nodes are down` | Retry after some time. |  
| User sends request after Node Hibernator has encountered an issue during hibernation/waking up of Ethereum Client or Privacy Manager | 500 (Internal Server Error) - `node is not ready to accept request` | Investigate the cause of Node Hibernator's failure and fix the issue. |  

*Note: Node Hibernator will consider a peer to be hibernated if it does not receive a response the peer's status during private transaction processing.*
