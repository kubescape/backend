# OperatorCommand

The OperatorCommand CRD is designed to enable the execution of various actions within the cluster and reporting their status back to the backend. This CRD serves as a central mechanism for triggering and managing actions, replacing the functionality previously provided by the gateway and kollector.

How it Works

1. Creation: The backend creates a Command CRD instance, specifying the desired action and any necessary parameters for the action.
2. Synchronization: The Synchronizer, responsible for two-way communication, receives the Command CRD from the backend and saves it in the cluster.
3. Execution: The designated component in the cluster, identifies the new command via a watcher on the Kubernetes API, processes the Command CRD and performs the requested action within the cluster.
4. Status Reporting: Upon completion, the component updates the command CRD resource with the status of the action, providing information about success or failure, any relevant details, and potentially updating the Command CRD. The synchronizer, watching over the command CRD, will send it back to the backend for further processing.
