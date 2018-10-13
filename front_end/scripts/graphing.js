
const TICKS = 5;
function formatData(data) {

	// Setting local time			
	Object.values(data).forEach(function(d) {
		d.forEach(function(i) {
			i.time = new Date(i.time);
		})
	})
	// if networth add points in to make a step graph
	return data;
	
};

// TODO: tags for d3 plotting(title labels etc) sent with dat object in an serparate property
//			tags can pass the type of data being sent through so more data structuring can be done here like min an maxs 
function DrawLineGraph(location, data, id) {
	// Pulling out data, use tags to change data if need
	var dat = formatData(data.data);
	var tags = data.tags;

	// logging remove later
	console.log("DATA");
	console.log(dat);
	console.log("TAGS");
	console.log(data.tags);
	// logging remove later

	var width = 700;
	var height = 500;
	var margin = {
		'top': 60,
		'bottom': 60,
		'left': 60,
		'right': 60,
	};
	console.log(location);
	var svg = d3.select(location).append('svg')
		.attr('width', width)
		.attr('height', height)
		.append("g");
	
	var minTime = new Date('3000 Jan 1');
	var maxTime = new Date('1999 Jan 1');
	var maxValue = Number.NEGATIVE_INFINITY;
	var minValue = Number.POSITIVE_INFINITY;
	var minValueTime, maxValueTime;
	for (var line_key in dat) {

		dat[line_key].forEach(function(d) {
			if (minValue > d.value) {
				minValue = d.value;
				minValueTime = d.time;
			}
			if (maxValue < d.value) {
				maxValue = d.value;
				maxValueTime = d.time;
			}
			if (minTime > d.time) {
				minTime = d.time;
			}
			if (maxTime < d.time) {
				maxTime = d.time;
			}
		});
	}


	// Creating graph scales
	var scaleTime = d3.scaleTime()
		.domain([minTime, maxTime])
		.range([margin.right, width - margin.left])
	var scaleValue = d3.scaleLinear()
		.domain([minValue - (minValue/10), maxValue + (maxValue/10)])
		.range([height  - margin.top, margin.bottom]);

		
	// Y grid
	function yGrid() {		
		return d3.axisLeft(scaleValue)
			.ticks(TICKS);
	}

	var labels = [];

	for (line_key in dat) {
		const timeFormat = d3.timeFormat("%d-%b-%y")
		
		var minY = Number.POSITIVE_INFINITY;
		var maxY = Number.NEGATIVE_INFINITY;
		var minX, maxX;
		dat[line_key].forEach(function(d) {
			if (minY > d.value) {
				minY = d.value;
				minX = d.time;
			}
			if (maxY < d.value) {
				maxY = d.value;
				maxX = d.time;
			}
		});

		//Add annotations
		var newLabels = [{
			data: { date: minX, value: formatPrice(minY) },
				x: scaleTime(minX),
				y: scaleValue(minY).toFixed(2),
				dx: 10,
				dy: 10
			}, {
			data: { date: maxX, value: formatPrice(maxY) },
				x: scaleTime(maxX),
				y:scaleValue(maxY).toFixed(2),	
				dx: 10,
				dy: 10
		}].map(function (l) {
			l.note = Object.assign({}, l.note, {
				title: "$" + l.data.value,
				label: "" + timeFormat(l.data.date) 
			});
			l.connector = {
				end: "dot",
			};

			return l;
		});

		newLabels.forEach(d => labels.push(d));

		let path = svg.append('path');

		let line = d3.line()
			.x(function(d) { return scaleTime(d.time); })
			.y(function(d) { return scaleValue(d.value); });

		// Adding line 
		path.data([dat[line_key]]).attr('d', line).attr('stroke', 'black').attr('stroke-width', '2px').attr('fill', 'none');
	}

	// Creating axis
	var xAxisCall = d3.axisBottom()
		.tickFormat(d3.timeFormat("%a %I:%M%p"))
		.ticks(TICKS);
	xAxisCall.scale(scaleTime);
	var yAxisCall = d3.axisLeft()
		.tickFormat(function(d) {
			return "$" + abbrevPrice(d);
		})
		.ticks(TICKS);
	yAxisCall.scale(scaleValue);

	// Add axis
	svg.append('g')
		.attr('id', 'x-axis')
		.attr('transform', 'translate(0, '+ (height - (margin.top-5)) +')')
		.style('font-size', 12)
		.call(xAxisCall)
	    .selectAll("text")	
	        .style("text-anchor", "end")
	        .attr("transform", "rotate(-35)")
	        .attr('font-size', '10px');

	svg.append('g')
		.attr('id', 'y-axis')
		.attr('transform', 'translate(' + (margin.left-5) +', ' + '0' + ')')
		.style('font-size', 12)
		.call(yAxisCall);

	// Add gridlines
	
	// add the Y gridlines
	svg.append("g")			
		.attr("class", "graph-grid")
		.call(yGrid()
			.tickSize(-width)
			.tickFormat("")
		)

	// Add graph title
	if (tags) {

		if (tags.title) {
			svg.append('text').text(tags.title)
				.attr('class', 'stockGraph-title')
				.attr('font-size', '20px')
				.attr('text-anchor', 'middle')
				.attr('transform', 'translate(' + (width/2) + ', 40)');
		}
	}
	console.log(labels)

	var annotate = d3.annotation().annotations(labels).editMode(true)
		.type(d3.annotationCallout)
		.accessors({ 
			x: d => scaleTime(d.date), 
			y: d => scaleValue(d.value).toFixed(2)
		  })
		.accessorsInverse({
			date: d => scaleTime.invert(d.x),
			value: d => scaleValue.invert(d.y).toFixed(2) 
		});

	// Adding annotations for extreme values
	svg.append('g')
		.attr('class', 'graph-annotation')
		.call(annotate);

};
	

	// // Adding axis labels
	// d3.select('#x-axis').append("text")
	// 	.attr('transform', 'translate(' + (margin.left/2) + ', ' + (margin.top/2) + ')')
	// 	.attr('fill', 'black')
	// 	.attr('font-size', 14)
	// 	.attr('font-weight', 'bold')
	// 	.text('Time');

	// d3.select('#y-axis').append('text')
	// 	.attr('transform', 'translate(0, ' + (margin.top/2) + ')')
	// 	.attr('fill', 'black')
	// 	.attr('font-size', 15)
	// 	.attr('font-weight', 'bold')
	// 	.text('$');


$('#select-timeframe').selectize({
    create: true,
    sortField: 'text'
});
