function runCommand(action)
{
	var xmlhttp;
	var shaObj = new jsSHA("SHA-1", "TEXT");
	var timestamp = (Math.floor(new Date() / 30000)).toString();
	
	var secret = document.getElementById('secret').value;

	shaObj.setHMACKey(secret, "TEXT");
	shaObj.update(timestamp);
	
	var hmac = shaObj.getHMAC("HEX");

	console.log("HMAC: " + hmac);

	if (window.XMLHttpRequest) {
		xmlhttp = new XMLHttpRequest();
		xmlhttp.open("GET", "/command/" + action, true);
		xmlhttp.setRequestHeader("Authorization", "Bearer " + hmac);
		xmlhttp.onreadystatechange = function() {
			if(xmlhttp.readyState === XMLHttpRequest.DONE)
			{
				let msg = xmlhttp.status + " - " + xmlhttp.responseText;

				console.log(msg);

				let info = document.getElementById('info');
				info.innerHTML = msg;

				setTimeout(function() {
					info.innerHTML = "Click a command to run it.";
				}, 3000);

			}
		};
		xmlhttp.send(null);
	}

}
	
function activate()
{
	runCommand("activate");

    var info = document.getElementById('info');
	info.innerHTML = "Activating door...";

}

function shutdown()
{
	var info = document.getElementById('info');
	info.innerHTML = "Shutting down Pi...";

	runCommand("shutdown");	
}

function reboot()
{
	var info = document.getElementById('info');
	info.innerHTML = "Rebooting Pi...";

	runCommand("reboot");
}

function lightson()
{
	var info = document.getElementById("info");
	info.innerHTML = "Lights on...";

	runCommand("lightson");
}

function lightsoff()
{
	var info = document.getElementById("info");
	info.innerHTML = "Lights off...";

	runCommand("lightsoff");
}
