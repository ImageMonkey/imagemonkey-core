var pollChartConfig = {
  type: 'pie',
  data: {
    datasets: [{
      data:  [0, 0],
      backgroundColor: ["#f40006", "#03bd4d"]
    }],
    labels: ["no", "yes"]
  },
  options: {
    responsive: true,
    title: {
      display: true,
      text: 'What other people are thinking'
    },
    legend: {
      display: true,
      position: 'bottom'
    }
  }
};

function isPollDataEmpty(){
  var d = pollChartConfig["data"]["datasets"][0]["data"];
  if((d[0] === 0) && (d[1] === 0))
    return true;
  return false;
}

function setPollData(numOfYes, numOfNo){
  pollChartConfig["data"]["datasets"][0]["data"] = [numOfNo, numOfYes];
}

function createPoll(){
  var pollChartCtx = $("#poll-chart")[0].getContext("2d");
  window.pollChart = new Chart(pollChartCtx, pollChartConfig);
}

function showHidePoll(show){
  if(show){
    $('#poll-chart').show();
  }
  else{
    $('#poll-chart').hide();
  }
}