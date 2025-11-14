const inputs = input.config();
await triggerLinkUpdate(inputs.recordIds[0], inputs.names[0], inputs.linkIds[0]);

async function triggerLinkUpdate(recordId, name, linkId) {
    if (!linkId) {
        console.error(`No short link ID found for record '${recordId}'; aborting`);
        return;
    }
    console.log(`Record id is '${recordId}'; client name is '${name}'`)
    const webhook = 'https://hook.us1.make.com/h61316aga8y1rjfwkweqkz6rpjqtbo81';
    const apiKey = 'cugXQDNfMdsfjmjAgngxumqWs';
    console.log(`Requesting reset of short link '${linkId}'...`);
    const response = await fetch(webhook, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-make-apikey': apiKey
        },
        body: JSON.stringify({
            recordId: recordId,
            shortLinkId: linkId
        })
    }).catch((error) => {
        console.error("Error sending request to make:", error);
    })
    if (response) {
        console.log(`Response code from make is ${response.status}`);
    }
}
