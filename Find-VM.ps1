function Find-VM{

param
([string]$ComputerName,
$credentials)

#List of SCVMM servers to query
$servers = Import-Csv -Path C:\users\jgreene\Documents\ToolBelt\SCVMM_Servers.csv

#Loop through servers and search for machine on each
foreach($server in $servers){   

#Capture server and VM info
$scvmmserv = Get-SCVMMServer $server.Server -Credential $credentials
$VM = Get-SCVirtualMachine -All -VMMServer $scvmmserv | Where-Object{$_.ComputerName -eq "$ComputerName.osisoft.int" -or $_.ComputerName -eq "$computerName.dev.osisoft.int"}

#Once found return info
if($VM.Name){Return $VM} 

}
#NO VM FOUND 
Return "No VM"
}