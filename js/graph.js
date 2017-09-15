var width = 960,
    height = 500,
    padding = 1.5, // separation between same-color nodes
    clusterPadding = 6, // separation between different-color nodes
    maxRadius = 50;

//get some predefined colors
var color = d3.scale.category20();

d3.text("", function(error, text) {
  if (error) throw error;
  var data = graphData;

  data.forEach(function(d) {
    d.size = +d.size;
  });


//unique cluster/group id's
//and min/max size
var cs = [];
data.forEach(function(d){
    if(!cs.contains(d.group)) {
        cs.push(d.group);
    }
});

var n = data.length, // total number of nodes
    m = cs.length; // number of distinct clusters

//create clusters and nodes
var clusters = new Array(m);
var nodes = [];
for (var i = 0; i<n; i++){
    nodes.push(create_nodes(data,i));
}

var force = d3.layout.force()
    .nodes(nodes)
    //.size([width, height])
    .gravity(.02)
    .charge(0)
    .on("tick", tick)
    .start();

var svg = d3.select("#graph").append("svg")
    .attr("width", '100%')
    .attr("height", '100%')
    .attr('viewBox','0 0 '+Math.min(width,height) +' '+Math.min(width,height) )
    .attr('preserveAspectRatio','xMinYMin')
    .append("g")
    .attr("transform", "translate(" + Math.min(width,height) / 2 + "," + Math.min(width,height) / 2 + ")");
    //.attr("width", width)
    //.attr("height", height)
    //.call(responsivefy);


var node = svg.selectAll("circle")
    .data(nodes)
    .enter().append("g").call(force.drag);


node.append("circle")
    .style("fill", function(d,i){return color(i);})
    .attr("r", function(d){return d.radius})
    

node.append("text")
      .attr("dy", ".3em")
      .style("text-anchor", "middle")
      .text(function(d) { return d.text.substring(0, d.radius / 3); });




function create_nodes(data,node_counter) {
  var i = cs.indexOf(data[node_counter].group),
      r = Math.sqrt((i + 1) / m * -Math.log(Math.random())) * maxRadius,
      d = {
        cluster: i,
        radius: data[node_counter].size*1.5,
        text: data[node_counter].text,
        x: Math.cos(i / m * 2 * Math.PI) * 200 + width / 2 + Math.random(),
        y: Math.sin(i / m * 2 * Math.PI) * 200 + height / 2 + Math.random()
      };
  if (!clusters[i] || (r > clusters[i].radius)) clusters[i] = d;
  return d;
};

function responsivefy(svg) {
    // get container + svg aspect ratio
    var container = d3.select(svg.node().parentNode),
        width = parseInt(svg.style("width")),
        height = parseInt(svg.style("height")),
        aspect = width / height;

    // add viewBox and preserveAspectRatio properties,
    // and call resize so that svg resizes on inital page load
    svg.attr("viewBox", "0 0 " + width + " " + height)
        .attr("perserveAspectRatio", "xMinYMid")
        .call(resize);

    // to register multiple listeners for same event type, 
    // you need to add namespace, i.e., 'click.foo'
    // necessary if you call invoke this function for multiple svgs
    // api docs: https://github.com/mbostock/d3/wiki/Selections#on
    d3.select(window).on("resize." + container.attr("id"), resize);

    // get width of container and resize svg to fit it
    function resize() {
        var targetWidth = parseInt(container.style("width"));
        svg.attr("width", targetWidth);
        svg.attr("height", Math.round(targetWidth / aspect));
    }
}



function tick(e) {
    node.each(cluster(10 * e.alpha * e.alpha))
        .each(collide(.5))
    .attr("transform", function (d) {
        var k = "translate(" + d.x + "," + d.y + ")";
        return k;
    })

}

// Move d to be adjacent to the cluster node.
function cluster(alpha) {
    return function (d) {
        var cluster = clusters[d.cluster];
        if (cluster === d) return;
        var x = d.x - cluster.x,
            y = d.y - cluster.y,
            l = Math.sqrt(x * x + y * y),
            r = d.radius + cluster.radius;
        if (l != r) {
            l = (l - r) / l * alpha;
            d.x -= x *= l;
            d.y -= y *= l;
            cluster.x += x;
            cluster.y += y;
        }
    };
}

// Resolves collisions between d and all other circles.
function collide(alpha) {
    var quadtree = d3.geom.quadtree(nodes);
    return function (d) {
        var r = d.radius + maxRadius + Math.max(padding, clusterPadding),
            nx1 = d.x - r,
            nx2 = d.x + r,
            ny1 = d.y - r,
            ny2 = d.y + r;
        quadtree.visit(function (quad, x1, y1, x2, y2) {
            if (quad.point && (quad.point !== d)) {
                var x = d.x - quad.point.x,
                    y = d.y - quad.point.y,
                    l = Math.sqrt(x * x + y * y),
                    r = d.radius + quad.point.radius + (d.cluster === quad.point.cluster ? padding : clusterPadding);
                if (l < r) {
                    l = (l - r) / l * alpha;
                    d.x -= x *= l;
                    d.y -= y *= l;
                    quad.point.x += x;
                    quad.point.y += y;
                }
            }
            return x1 > nx2 || x2 < nx1 || y1 > ny2 || y2 < ny1;
        });
    };
}
});

Array.prototype.contains = function(v) {
    for(var i = 0; i < this.length; i++) {
        if(this[i] === v) return true;
    }
    return false;
};