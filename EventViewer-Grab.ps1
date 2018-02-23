
function GetEventsNShit {

        $chromeCache='%LocalAppData%\Google\Chrome\User Data\Default\Cache'
        $localhost = 'localhost'

param( [string]$computerName = $localhost ,[datetime]$startTime=(Get-Date).AddMinutes(-15) ,[datetime]$Endtime=(Get-DAte))

$logs = (Get-WinEvent -ListLog * -ComputerName $computerName | Where-Object{$_.recordCount}).Logname

$filterTable=@{
        'StartTime' = $startTime
        'EndTime' = $Endtime
        'LogName' = $logs
}

$events = Get-WinEvent -ComputerName $computerName -FilterHashtable $filterTable -ErrorAction SilentlyContinue

Return $events


}

