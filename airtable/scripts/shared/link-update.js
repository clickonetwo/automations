async function triggerLinkUpdate(hook, recordId, linkId) {
    console.log(`Requesting reset of link '${linkId}' from record ${recordId}...`);
    const response = await fetch(hook, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
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
