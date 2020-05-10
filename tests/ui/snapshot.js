/**
 * small script which creates a screenshot of the annotation with the provided annotation id
 *
 * Usage:
 *   node snapshot.js <annotation id>
 *
 *   e.g:
 *
 *   node snapshot.js f2164814-507c-4760-9f30-d05250938506 
 *
 **/

const puppeteer = require('puppeteer');

var annotationId = process.argv[2];
if(annotationId === undefined) {
	console.log('Please specify a annotation id!');
	return;
}

const waitTillHTMLRendered = async (page, timeout = 30000) => {
  const checkDurationMsecs = 1000;
  const maxChecks = timeout / checkDurationMsecs;
  let lastHTMLSize = 0;
  let checkCounts = 1;
  let countStableSizeIterations = 0;
  const minStableSizeIterations = 3;

  while(checkCounts++ <= maxChecks){
    let html = await page.content();
    let currentHTMLSize = html.length; 

    let bodyHTMLSize = await page.evaluate(() => document.body.innerHTML.length);

    console.log('last: ', lastHTMLSize, ' <> curr: ', currentHTMLSize, " body html size: ", bodyHTMLSize);

    if(lastHTMLSize != 0 && currentHTMLSize == lastHTMLSize) 
      countStableSizeIterations++;
    else 
      countStableSizeIterations = 0; //reset the counter

    if(countStableSizeIterations >= minStableSizeIterations) {
      console.log("Page rendered fully..");
      break;
    }

    lastHTMLSize = currentHTMLSize;
    await page.waitFor(checkDurationMsecs);
  }  
};

(async () => {
  const browser = await puppeteer.launch({args: ['--no-sandbox', '--disable-setuid-sandbox']});
  const page = await browser.newPage();
  await page.goto('http://127.0.0.1:8080/annotate?view=unified&annotation_id='+annotationId);
  
  await waitTillHTMLRendered(page);
  
  await page.screenshot({path: annotationId+'.png'});

  await browser.close();
})();
