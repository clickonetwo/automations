async function triggerLinkUpdate(table, record) {
    const nameField = table.getField('fldGF8G0cEoxqKgrd'); // Name
    const formLinkField = table.getField('fld73xz3hRC09PTQZ'); // Client Form URL
    const shortLinkField = table.getField('fld68qXdAoFhzQCBF'); // Client Short Link
    const shortLinkIdField = table.getField('fldiXUviO9JSM5yr7'); // Client Short Link ID
    const id = record.id;
    const name = record.getCellValueAsString(nameField);
    const formLink = record.getCellValueAsString(formLinkField);
    const shortLink = record.getCellValueAsString(shortLinkField);
    const shortLinkId = record.getCellValueAsString(shortLinkIdField);
    console.log(`Record id is ${record.id}; client name is ${name}`)
    const webhook = 'https://hook.us1.make.com/2diez3ugn1hlsw1ophpxh96wqw35hq83';
    const apiKey = 'cugXQDNfMdsfjmjAgngxumqWs';
    if (shortLinkIdField) {
        console.log(`Requesting creation of new short link...`);
    } else {
        console.log(`Requesting reset of existing short link...`);
    }
    const response = await fetch(webhook, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-make-apikey': apiKey
        },
        body: JSON.stringify({
            "recordId": id,
            "name": name,
            "formLink": formLink,
            "shortLink": shortLink,
            "shortLinkId": shortLinkId
        })
    }).catch((error) => {
        console.error("Error sending request to make:", error);
    })
    console.log(`Response code from make is ${response.status}`)
}
