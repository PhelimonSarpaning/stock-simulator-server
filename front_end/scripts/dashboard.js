// TODO break these out another file
var REQUESTS = {};
var REQUEST_ID = 1;


// Get saved data from sessionStorage
$( document ).ready(function() {

	/* Highest level Vue data object */
	var config = new Vue({
		data: {
			config: {},
		}
	})

	var vm_stocks = new Vue({
		data: {
			stocks: {},
		}
	});

	var vm_ledger = new Vue({
		data: {
			ledger: {},
		}
	});

	var vm_portfolios = new Vue({
		data: {
			portfolios: {},
		}
	});

	var vm_users = new Vue({
		data: {
		  users: {},
		  currentUser: auth_uuid,
		},
		methods: {
	  		getCurrentUser: function() {
	  			// Get userUUID of the person that is logged in
	  			var currentUser = sessionStorage.getItem('uuid');
	  			if (currentUser !== null) {
		  			// Have they been added to the users object yet?
		  			if (vm_users.users[currentUser]) {
		  				return vm_users.users[currentUser].display_name;
		  			} else {
		  				return "";
		  			}
	  			}
	  		}
	  	},
	});

	console.log("----- CONFIG -----");
	console.log(config.config);
	console.log("----- USERS -----");
	console.log(vm_users.users);
	console.log("------ STOCKS ------");
	console.log(vm_stocks.stocks);
	console.log("------ LEDGER ------");
	console.log(vm_ledger.ledger);
	console.log("------ PORTFOLIOS ------");
	console.log(vm_portfolios.portfolios);



	/* Vues that are used to display data */

	// Vue for sidebar navigation
	let vm_nav = new Vue({
		el: '#nav',
		methods: {
			nav: function (event) {

				let route = event.currentTarget.getAttribute('data-route');
				
				renderContent(route);
		    }
		}
	});

	// Vue for any sidebar data
	var sidebarCurrUser = new Vue({
		el: '#stats--view',
		methods: {
			toPrice: formatPrice,
		},
		computed: {
			currUserPortfolio: function() {
				var currUserUUID = sessionStorage.getItem('uuid');
				if (vm_users.users[currUserUUID] !== undefined) {
	    			var currUserFolioUUID = vm_users.users[currUserUUID].portfolio_uuid;
					if (vm_portfolios.portfolios[currUserFolioUUID] !== undefined) {
		    			var folio = vm_portfolios.portfolios[currUserFolioUUID];
						folio.investments = folio.net_worth - folio.wallet;
		    			return folio;
	    			}
	    		}
	    		return {};
	    	}
		}
	});

	// Vue for username top right
	var topBar = new Vue({
		el: "#top-bar--container",
		methods: {
			logout: function (event) {
				// delete cookie
				// Get saved data from sessionStorage
				console.log("logout");
				sessionStorage.removeItem('token');
				sessionStorage.removeItem('auth_obj');
				// send back to index
				window.location.href = "/login.html";
		    },
		    changeDisplayName: function() {
		    	// Get entered display name
		    	let new_name = $("#newDisplayName").val();

		    	// Creating message that changes the users display name
				let msg = {
					'set': "display_name",
					'value': new_name
				};

				REQUESTS[REQUEST_ID] = function(msg) {
					alert("Display_name changed to: " + new_name);
				};
				// Send through WebSocket
				console.log(JSON.stringify(msg));
		    	doSend('set', msg, REQUEST_ID.toString());

		    	REQUEST_ID++;

		    	// Reset display name
		    	$("#newDisplayName").val("");
		    }
		},
		computed: {
			userDisplayName: function() {
				var currUserUUID = sessionStorage.getItem('uuid');
				if (vm_users.users[currUserUUID] !== undefined) {
	    			return vm_users.users[currUserUUID].display_name;
	    		}
	    		return "";
			}
		}
	});

	// Vue for all dashboard data
	var vm_dash_tab = new Vue({
		el: '#dashboard--view',
		data: {
		  sortBy: 'amount',
		  sortDesc: 1,
		},
		methods: {
			toPrice: formatPrice,
			// on column name clicks
		    sortCol: function(col) {
				// If sorting by selected column
		    	if (vm_dash_tab.sortBy == col) {
					// Change sort direction
		    		// console.log(col);
		    		vm_dash_tab.sortDesc = -vm_dash_tab.sortDesc;
		    	} else {
					// Change sorted column
		    		vm_dash_tab.sortBy = col;
		    	}
		    },
		    createPortfolioGraph: function() {
				// Get curr user portfolioUUID
	    		let portfolioUUID = vm_dash_tab.currUserPortfolio.uuid;
	    		let location = "#portfolio-graph";
	    		createPortfolioGraph(portfolioUUID, location);
	    	}
		},
		computed: {
			currUserPortfolio: function() {
				var currUserUUID = sessionStorage.getItem('uuid');
				if (vm_users.users[currUserUUID] !== undefined) {
					var currUserFolioUUID = vm_users.users[currUserUUID].portfolio_uuid;
					if (vm_portfolios.portfolios[currUserFolioUUID] !== undefined) {
						var folio = vm_portfolios.portfolios[currUserFolioUUID];
						folio.investments = folio.net_worth - folio.wallet;
		    			return folio;
	    			}
	    		}
	    		return {};
	    	},
			currUserStocks: function() {
				var currUserUUID = sessionStorage.getItem('uuid');
				if (vm_users.users[currUserUUID] !== undefined) {
					
					// Current users portfolio uuid
					var portfolio_uuid = vm_users.users[currUserUUID].portfolio_uuid;
					
					// If objects are in ledger
					if (Object.keys(vm_ledger.ledger).length !== 0) {
						
						var ownedStocks = Object.values(vm_ledger.ledger).filter(d => d.portfolio_id === portfolio_uuid);
						
						// Remove stocks that user owns 0 of
						ownedStocks = ownedStocks.filter(d => d.amount !== 0);
						// Augmenting owned stocks
						ownedStocks = ownedStocks.map(function(d) {
							d.stock_ticker = vm_stocks.stocks[d.stock_id].ticker_id;
							d.stock_price = vm_stocks.stocks[d.stock_id].current_price;
							d.stock_value = Number(d.stock_price) * Number(d.amount);								
							d.stock_roi = (Number(d.stock_price) * Number(d.amount)) - Number(d.investment_value);

							// TODO: css changes done here talk to brennan about his \ux22 magic 
							// helper to color rows in the stock table 
							var targetChangeElem = $("tr[uuid=\x22" + d.stock_uuid + "\x22].clickable > td.stock-change");
							// targetChangeElem.addClass("rising");
							// if (d.stock_roi > 0) {
							// 	targetChangeElem.removeClass("falling");
							// 	targetChangeElem.addClass("rising");
							// } else if (d.stock_roi === 0) {
							// 	targetChangeElem.removeClass("falling");
							// 	targetChangeElem.removeClass("rising");
							// } else {
							// 	targetChangeElem.removeClass("rising");
							// 	targetChangeElem.addClass("falling");
							// }
							return d;
						});

				    	// Sorting array of owned stocks
				    	ownedStocks = ownedStocks.sort(function(a,b) {
				    		if (a[vm_dash_tab.sortBy] > b[vm_dash_tab.sortBy]) {
				    			return -vm_dash_tab.sortDesc;
				    		} 
				    		if (a[vm_dash_tab.sortBy] < b[vm_dash_tab.sortBy]) {
				    			return vm_dash_tab.sortDesc;
				    		}
				    		return 0;
				    	});
						return ownedStocks;
					}
				}
				return [];
			},
		}
	});


	// Vue for all stocks tab data 
	var vm_stocks_tab = new Vue({
		el: '#stocks--view',
		data: {
		  sortBy: 'ticker_id',
		  sortDesc: 1,
		  sortCols: ["ticker_id", "open_shares", "change", "current_price"],
		  sortDirections: [-1, -1, -1, -1],
		  reSort: 1,
		},
		methods: {
			toPrice: formatPrice,
		    // on column name clicks
		    sortCol: function(col) {
				// If sorting by selected column
		    	if (vm_stocks_tab.sortBy == col) {
					// Change sort direction
		    		// console.log(col);
		    		vm_stocks_tab.sortDesc = -vm_stocks_tab.sortDesc;
		    	} else {
					// Change sorted column
		    		vm_stocks_tab.sortBy = col;
		    		
		    	}
		    },
		    multiSort: function(col) {
		    	// if old first sort is the new first sort
		    	if (vm_stocks_tab.sortCols[0] === col) {
		    		// change sort direction
		    		vm_stocks_tab.sortDirections[0] *= -1;
		    	} else {
		    		// Where is the new sort column
		    		let ind = vm_stocks_tab.sortCols.indexOf(col);
		    		// Remove new column from old spot
		    		vm_stocks_tab.sortCols.splice(ind, 1);
		    		vm_stocks_tab.sortDirections.splice(ind, 1);
		    		// Push to the beginning of the array
		    		vm_stocks_tab.sortCols.unshift(col);
		    		vm_stocks_tab.sortDirections.unshift(1);
		    	}
		    	vm_stocks_tab.reSort++;
		    },
		    createStockGraph: function(stockUUID) {
		    	let stock = Object.values(vm_stocks.stocks).filter(d => d.uuid === stockUUID)[0];
		    	console.log(stock)
		    	
				// Creating data object and adding tags
				let data = {
		    		data: {},
		    		tags: {
		    			'title': stock.ticker_id,
		    			'type': 'stock',
		    		},
		    	};
		    	// Creating message 
					let msg = {
							'uuid': stockUUID,
							'field': 'current_price',
							'num_points': 100,
							'length': "100h",
						};
					
					// Store request on front end
					REQUESTS[REQUEST_ID] = function(msg) {
						// Pull out the data and format it
						var points = msg.msg.points;
						points = points.map(function(d) {
							return {'time': d[0], 'value': d[1] };
						});

						// Store the data
						data.data[stockUUID] = points;

						// Make note the data is available
						DrawLineGraph('#stock-list', data);
					};
					
					// Send message
					doSend('query', msg, REQUEST_ID.toString());

					REQUEST_ID++;

		    },
		    createStocksGraph: function() {
		    	console.log("Creating stock graphs");
		    	// Store graphing data
				var data = {
					data: {},
					tags: {},
				};
				var responses = [];
				var requests = [];

				// Send data requests
				Object.keys(vm_stocks.stocks).forEach(function(stockUUID) {
					vm_stocks_tab.createStockGraph(stockUUID);
				});

				Object.keys(vm_stocks.stocks).forEach(function(stockUUID) {
					// Creating message 
					let msg = {
						'uuid': stockUUID,
						'field': 'current_price',
						'num_points': 100,
						'length': "100h",
					};
					
					// Store request on front end
					requests.push(REQUEST_ID.toString());
					REQUESTS[REQUEST_ID] = function(msg) {
						// Pull out the data and format it
						var points = msg.msg.points;
						points = points.map(function(d) {
							return {'time': d[0], 'value': d[1] };
						});

						// Store the data
						data.data[msg.msg.message.uuid] = points;

						// Make note the data is available
						responses.push(msg.request_id);
						// addToLineGraph('#portfolio-graph', points, field);
					};

					// Send message
					doSend('query', msg, REQUEST_ID.toString());

					REQUEST_ID++;

				});

				var drawGraphOnceDone = null

				var stillWaiting = true;
				
				drawGraphOnceDone = function(){
					if (requests.every(r => responses.indexOf(r) > -1)) {
						stillWaiting = false;
					}

					if (!stillWaiting) {
				    	DrawLineGraph('#stock-graph', data);
					} else {
						setTimeout(drawGraphOnceDone, 100);
					}
				}

				setTimeout(drawGraphOnceDone, 100);

		    },
		},
		computed: {
			sortedStocks: function() {
	    		if (Object.keys(vm_stocks.stocks).length !== 0) {
		    	  	// Turn to array and sort 
					var stock_array = Object.values(vm_stocks.stocks);

			    	// Sorting array
			    	stock_array = stock_array.sort(function(a,b) {
			    		if (a[vm_stocks_tab.sortBy] > b[vm_stocks_tab.sortBy]) {
			    			return -vm_stocks_tab.sortDesc;
			    		}
			    		if (a[vm_stocks_tab.sortBy] < b[vm_stocks_tab.sortBy]) {
			    			return vm_stocks_tab.sortDesc;
			    		}
			    		return 0;
			    	})
			    	return stock_array;
				}
				return [];
			},
			multiSortStocks: function() {
				if (Object.keys(vm_stocks.stocks).length !== 0) {
					
					function sorter(a, b, ind) {
						if (a[vm_stocks_tab.sortCols[ind]] > b[vm_stocks_tab.sortCols[ind]]) {
							return vm_stocks_tab.sortDirections[ind];
						}
						if (a[vm_stocks_tab.sortCols[ind]] < b[vm_stocks_tab.sortCols[ind]]) {
							return -vm_stocks_tab.sortDirections[ind];
						}
						if (ind === (vm_stocks_tab.sortCols.length-1)) {
							return 0;
						} else {
							return sorter(a, b, ind+1);
						}
					};

					// Get all stocks
					var stock_array = Object.values(vm_stocks.stocks);
					// Sort
					stock_array = stock_array.sort(function(a,b) {
						return sorter(a, b, 0);
					});

					return stock_array;
				}
				return [];
			},
			highestStock: function() {
				if (Object.values(vm_stocks.stocks).length === 0) {
					return "";
				} else {
					stocks = Object.values(vm_stocks.stocks).map((d) => d);
					var highestStock = stocks.reduce((a, b) => a.current_price > b.current_price ? a : b);
					return highestStock.ticker_id;
				}
			},
			mostChange: function() {
				if (Object.values(vm_stocks.stocks).length === 0) {
					return "";
				} else {
					stocks = Object.values(vm_stocks.stocks).map((d) => d);
					var mover = stocks.reduce((a, b) => a.change > b.change ? a : b);
					return mover.ticker_id;
				}
			},
			lowestStock: function() {
				if (Object.values(vm_stocks.stocks).length === 0) {
					return "";
				} else {
					stocks = Object.values(vm_stocks.stocks).map((d) => d);
					var mover = stocks.reduce((a, b) => a.current_price < b.current_price ? a : b);
					return mover.ticker_id;
				}
			},
		}
	});

	// Vue for all investors tab data
	var vm_investors_tab = new Vue({
		el: '#investors--view',
		methods: {
			toPrice: formatPrice,
			createGraph: function(portfolioUUID) {
				let location = "#investorGraph" + portfolioUUID;
				createPortfolioGraph(portfolioUUID, location);
			}
		},
		computed: {
			investors: function() {
				var investors = Object.values(vm_portfolios.portfolios);
				// List of all ledger items
				var ledgerItems = Object.values(vm_ledger.ledger);

				investors.map(function(d) {
					// Augment investor data
					d.name = vm_users.users[d.user_uuid].display_name;
					// Get all stocks
					d.stocks = ledgerItems.filter(l => (l.portfolio_id === d.uuid) & (l.amount !== 0)); // ledgers can have amount == 0
					// Augment stock data
					d.stocks = d.stocks.map(function(d) {
						d.ticker_id = vm_stocks.stocks[d.stock_id].ticker_id;
						d.stock_name = vm_stocks.stocks[d.stock_id].name;
						d.current_price = vm_stocks.stocks[d.stock_id].current_price;
						d.value = d.current_price * d.amount;
						return d;
					});

					return d;
				});
				return investors;
			},
		}
	});

	function createPortfolioGraph(portfolioUUID, location) {
		// Store graphing data
		var data = {
			data: {},
			tags: {},
		};
		var responses = [];
		var requests = [];

		// Send data requests
		["net_worth", "wallet"].forEach(function(field) {

			// Creating websocket message 
			let msg = {
				'uuid': portfolioUUID,
				'field': field,
				'num_points': 100,
				'length': "100h",
			};

			// Store request on front end
			requests.push(REQUEST_ID.toString());
			REQUESTS[REQUEST_ID] = function(msg) {
				// Pull out the data and format it
				var points = msg.msg.points;
				points = points.map(function(d) {
					return {'time': d[0], 'value': d[1] };
				});

				// Store the data
				data.data[msg.msg.message.field] = points;

				// Make note the data is available
				responses.push(msg.request_id);
			};

			// Send message
			doSend('query', msg, REQUEST_ID.toString());

			REQUEST_ID++
		});

		var drawGraphOnceDone = null

		var stillWaiting = true;
		
		drawGraphOnceDone = function() {
			if (requests.every(r => responses.indexOf(r) > -1)) {
				stillWaiting = false;
			}

			if (!stillWaiting) {
				DrawLineGraph(location, data);
			} else {
				setTimeout(drawGraphOnceDone, 100);
			}
		}

		setTimeout(drawGraphOnceDone, 100);
	};

	Vue.component('investor-card', {
		computed: {
			investor: function() {
				return vm_investors_tab.investors.filter(d => d.user_uuid === this.user_uuid)[0];
			}
		},
		props: ['user_uuid'],
		template: "<div>{{ investor }}</div>"
	});

	var notification_sound = new Audio();
	notification_sound.src = "assets/sfx_pling.wav";

	var vm_chat = new Vue({
		el: '#chat-module--container',
		data: {
			showingChat: false,
			unreadMessages: false,
			mute_notification_sfx: false,
		},
		methods: {
			toggleChat: function() {
				this.showingChat = !this.showingChat;
				this.unreadMessages = false;
	    		$('#chat-module--container').toggleClass('closed');
	    		$('#chat-text-input').focus();
			},
			activeUsers: function() {
				// stop here later when not concating a string
				var online = Object.values(vm_users.users).filter(d => d.active === true);

				var online_str = JSON.stringify(online.map(d => d.display_name).join(', '));
				return online_str.replace(/"/g, "");
			}
		},
		computed: {
			numActiveUsers: function() {
				return Object.values(vm_users.users).filter(d => d.active === true).length;
			},
		},
		watch: {
			unreadMessages: function() {
				// make css changes here to show a notification for unread messages
				if (this.unreadMessages) {
					console.log("unread messages");
					$("#chat-module--container .chat-title-bar span").addClass("unread");
					if(!vm_chat.mute_notification_sfx) {
						notification_sound.play();
					}
				} else {
					console.log("all messages read");
					$("#chat-module--container .chat-title-bar span").removeClass("unread");
				}
			}
		},
	});


	$(document).scroll(function() {
		scrollVal = $(document).scrollTop();
	});


	let chat_feed = $('#chat-module--container .chat-message--list');
	let debug_feed = $('#debug-module--container .debug-message--list');

	var cleanInput = (input) => {
		return $('<div/>').text(input).html();
	}

	function appendNewMessage(msg, fromMe){
		// if your chat is closed, add notification
		if (!vm_chat.showingChat) {
			
			vm_chat.unreadMessages = true;
			
			if(!vm_chat.mute_notification_sfx) {
				notification_sound.play();
			}
			
		}

		let msg_text = cleanInput(msg.body);
		let msg_author_display_name = msg.author_display_name;
		let msg_author_uuid = msg.author_uuid;
		let msg_timestamp = formatDate12Hour(new Date($.now()));

		let msg_template = '';			
		let isMe = "";
		if (fromMe) {
			isMe = "is-me";
			msg_template = '<li '+ msg_author_uuid +'>'+
				'				<div class="msg-timestamp">'+ msg_timestamp +'</div>'+
				'				<div class="msg-username '+ isMe +'">'+ msg_author_display_name +'</div>'+
				'				<div class="msg-text right">'+ msg_text +'</div>'+
				'			</li>';
		} else {
			msg_template = '<li '+ msg_author_uuid +'>'+
				'				<div class="msg-timestamp">'+ msg_timestamp +'</div>'+
				'				<div class="msg-username '+ isMe +'">'+ msg_author_display_name +'</div>'+
				'				<div class="msg-text">'+ msg_text +'</div>'+
				'			</li>';
		}

		chat_feed.append(msg_template);
		chat_feed.animate({scrollTop: chat_feed.prop("scrollHeight")}, $('#chat-module--container .chat-message--list').height());

	}

	$('.debug-title-bar button').click(function() {

	    $('#debug-module--container').toggleClass('closed');
	    //$('#debug-text-input').focus();
	});

	$('.account-settings-btn').click(function() {
		// console.log("clicked");
	    $('#top-bar--container .account-settings-menu--container').toggleClass('open');
	    
	});

	$('#account-settings-menu-close-btn').click(function() {

	    $('#top-bar--container .account-settings-menu--container').toggleClass('open');
	    
	});

	$('.debug-btn').click(function() {

	    $('#debug-module--container').toggleClass('visible');
	    
	});

	$('#modal--container').click(function() {
		
		console.log("modal quit");
	    $('#modal--container').removeClass('open');
	    
	});


	// TODO: find a better spot for this
	$('table').on('click', 'tr.clickable' , function (event) {
		
		var ticker_id = $(this).find('.stock-ticker-id').attr('tid');
		
		console.log("TID: "+ticker_id);

		var stock = Object.values(vm_stocks.stocks).filter(d => d.ticker_id === ticker_id)[0];

		// Set show modal to true
		buySellModal.showModal = true;
		buySellModal.stock_uuid = stock.uuid;
	    
	    toggleModal();
	    
	});

	$('table').on('click', 'tr.investors' , function (event) {
		
		//var ticker_id = $(this).find('.stock-ticker-id').attr('tid');
		
		//console.log("TID: "+ticker_id);

		//var stock = Object.values(vm_users.stocks).filter(d => d.ticker_id === ticker_id)[0];

		// Set show modal to true
		genericModal.showModal = true;
		//genericModal.investor_uuid = stock.uuid;
	    
	    toggleGenericModal();
	    
	});


	$('thead tr th').click(function(event) {
		
		if($(event.currentTarget).find('i').hasClass("shown")) {
			$(event.currentTarget).find('i').toggleClass("flipped");
			// console.log("is asc");
		} else {
			$('thead tr th i').removeClass("shown");
			$(event.currentTarget).find('i').addClass("shown");
		}
	});


	function formatChatMessage(msg) {

		let timestamp = formatDate12Hour(new Date($.now()));
		// let message_body = $('#chat-module--container textarea').val();
		let message_body = msg.msg.message_body;
		let message_author = msg.msg.author;
		let isMe = false;

		if(vm_users.currentUser === message_author) {
			isMe = true;
			//console.log(isMe);
		} else {
			isMe = false;
		}

		let temp_msg = {
	        author_uuid: message_author,
	        author_display_name: vm_users.users[message_author].display_name,
	        timestamp: timestamp,
	        body: message_body,
	    };

	    appendNewMessage(temp_msg, isMe);
	}


	$(document).keypress(function(e) {
		
		if($('#chat-module--container textarea').val()) {
		    
		    if(e.which == 13) {

		    	let message_body = $('#chat-module--container textarea').val();
				
				var msg = {
					'message_body': message_body,
				};

				doSend('chat', msg)
		        
		        $('#chat-module--container textarea').val().replace(/\n/g, "");
		        $('#chat-module--container textarea').val('');
		        return false;
		    }
		}
	});


	$(document).keyup(function(e) {
	  	
	  	if(buySellModal.showModal === true){
	  		if (e.keyCode === 27) {
				//toggleModal();
				buySellModal.closeModal();
			}   
	  	}
		
	});



	// var routeConnect = function(msg) {
	registerRoute('connect', function(msg) {

	    console.log("login recieved");

	    if(!msg.msg.success) {
	        let err_msg = msg.msg.err;
	        console.log(err_msg);
	        console.log(msg);
	        window.location.href= "/login.html";
	    } else {
	        console.log(msg);
	        sessionStorage.setItem('uuid', msg.msg.uuid);
	    	Vue.set(config.config, msg.msg.uuid, msg.msg.config);
	    }
	});

	registerRoute('object', function(msg) {
		switch (msg.msg.type) {
			case 'portfolio':
				//console.log(msg.msg.object)
			    Vue.set(vm_portfolios.portfolios, msg.msg.uuid, msg.msg.object);
			    break;

			case 'stock':
		  		// Add variables for stocks for vue module initialization 
		  		msg.msg.object.change = 0;
		  		Vue.set(vm_stocks.stocks, msg.msg.uuid, msg.msg.object);
			  	break;
				  
			case 'ledger':
		  		Vue.set(vm_ledger.ledger, msg.msg.uuid, msg.msg.object);
			  	break;

			case 'user':
			    Vue.set(vm_users.users, msg.msg.uuid, msg.msg.object);
			    break;

		}
	});


	registerRoute('update', function(msg) {
		var updateRouter = {
			'stock': stockUpdate,
			'ledger': ledgerUpdate,
			'portfolio': portfolioUpdate,
			'user': userUpdate,
		};
		updateRouter[msg.msg.type](msg);
	});


	registerRoute('response', function(msg) {
		try {
			REQUESTS[msg.request_id](msg);
		} catch(err) {
			console.log(err);
			console.log("no request_id key for " + JSON.stringify(msg));
			console.log(REQUESTS);
			console.log(REQUEST_ID);
		}
		console.log(msg);
		delete REQUESTS[msg.request_id];
	});


	registerRoute('alert', function(msg) {
		console.log(msg);
	});


	registerRoute('chat', function(msg) {
		console.log("----- CHAT -----");
		console.log(msg);
		formatChatMessage(msg);
	});


	var stockUpdate = function(msg) {
		var targetUUID = msg.msg.uuid;
		msg.msg.changes.forEach(function(changeObject){
			// Variables needed to update the stocks
			var targetField = changeObject.field;
			var targetChange = changeObject.value;
			
			// if value to update is current price, calculate change
			if (targetField === "current_price") {
				// temp var for calculating price
				var currPrice = vm_stocks.stocks[targetUUID][targetField];
				// Adding change amount
				vm_stocks.stocks[targetUUID].change = targetChange - currPrice;
				// vm_stocks.stocks[targetUUID].change = Math.round((targetChange - currPrice) * 1000)/100000;
				
				// helper to color rows in the stock table 
				var targetElem = $("tr[uuid=\x22" + targetUUID + "\x22]");
				var targetChangeElem = $("tr[uuid=\x22" + targetUUID + "\x22] > td.stock-change");
				
				if ((targetChange - currPrice) > 0) {
					targetChangeElem.removeClass("falling");
					targetChangeElem.addClass("rising");
				} else {
					targetChangeElem.removeClass("rising");
					targetChangeElem.addClass("falling");
				}
			}

			// Adding new current price
			vm_stocks.stocks[targetUUID][targetField] = targetChange;

		})
	};

	var ledgerUpdate = function(msg) {

		var targetUUID = msg.msg.uuid;
		msg.msg.changes.forEach(function(changeObject){
			// Variables needed to update the ledger item
			var targetField = changeObject.field;
			var targetChange = changeObject.value;

			// Update ledger item
			vm_ledger.ledger[targetUUID][targetField] = targetChange;
		})
	};

	var portfolioUpdate = function(msg) {

		var targetUUID = msg.msg.uuid;
		msg.msg.changes.forEach(function(changeObject){
			// Variables needed to update the ledger item
			var targetField = changeObject.field;
			var targetChange = changeObject.value;

			// Update ledger item
			vm_portfolios.portfolios[targetUUID][targetField] = targetChange;
		})
	};

	var userUpdate = function(msg) {

		var targetUUID = msg.msg.uuid;
		msg.msg.changes.forEach(function(changeObject){
			// Variables needed to update the ledger item
			var targetField = changeObject.field;
			var targetChange = changeObject.value;

			// Update ledger item
			vm_users.users[targetUUID][targetField] = targetChange;
		})
	};



	/* Sending trade requests */

	function sendTrade() {
		
		// Creating message for the trade request
		var msg = {
			'stock_id': buySellModal.stock_uuid,
			'amount': buySellModal.buySellAmount,
		}

		// Sending through websocket
		console.log("SEND TRADE");

		REQUESTS[REQUEST_ID] = function(msg) {
			if (msg.msg.success) {
				notify("Trade successful!", msg.msg.success);
			} else {
				notify("Trade unsuccessful: " + msg.msg.err, msg.msg.success);
			}
		};

		// Send through WebSocket
		console.log(JSON.stringify(msg));
		doSend('trade', msg, REQUEST_ID.toString());
    	
    	REQUEST_ID++;

		// Reset buy sell amount
		buySellModal.buySellAmount = 0;

	};

		/* End sending trade requests */


	function toggleModal() {
		$('#modal--container').toggleClass('open');
	}

	function toggleGenericModal() {
		console.log("Show generic modal");
		$('#generic-modal--container').toggleClass('open');
	}


	// Vue object for the buy and sell modal
	var buySellModal = new Vue({
		el: '#modal--container',
		data: {
			showModal: false,
			buySellAmount: 0,
			isBuying: true,
			stock_uuid: 'OSRS',
		},
		methods: {
			toPrice: formatPrice,
			addAmount: function(amt) {
				buySellModal.buySellAmount += amt; 
			},
			clearAmount: function() {
				buySellModal.buySellAmount = 0;
			},
			determineMax: function() {
				if (buySellModal.isBuying) {
					buySellModal.buySellAmount = buySellModal.stock.open_shares; 
				} else {
					//determine current users holdings
					let stock = vm_dash_tab.currUserStocks.filter(d => d.stock_id === buySellModal.stock_uuid)[0];
					if (stock !== undefined) {
						buySellModal.buySellAmount = stock.amount;
					} else {
						buySellModal.buySellAmount = 0;
					}
				}
			},
			setIsBuying: function(bool) {
				// Change buying or selling
				buySellModal.isBuying = bool;

				// Set styling
				if (buySellModal.isBuying) {
		        	$('#calc-btn-buy').addClass("fill");
		        	$('#calc-btn-sell').removeClass("fill");
				} else {
		        	$('#calc-btn-sell').addClass("fill");
		        	$('#calc-btn-buy').removeClass("fill");
				}
			},
			submitTrade: function() {
				// Change amount depending on buy/sell
				if (!buySellModal.isBuying) {
					buySellModal.buySellAmount *= -1;
				}
				sendTrade();
				toggleModal();
			},
			closeModal: function(){
				toggleModal();
				buySellModal.buySellAmount = 0;
				buySellModal.showModal = false;
				buySellModal.isBuying = true;
			}
		},
		computed: {
			stock: function() {

				var clickedStock = Object.values(vm_stocks.stocks).filter(d => d.uuid === buySellModal.stock_uuid)[0];
				return clickedStock;
			},
			user: function() {
				var currUserUUID = sessionStorage.getItem('uuid');
				if (vm_users.users[currUserUUID] !== undefined) {
	    			var currUserFolioUUID = vm_users.users[currUserUUID].portfolio_uuid;
					if (vm_portfolios.portfolios[currUserFolioUUID] !== undefined) {
		    			var folio = vm_portfolios.portfolios[currUserFolioUUID];
						folio.investments = folio.net_worth - folio.wallet;
		    			return folio;
	    			}
	    		}
	    		return {};
			}
		},
		watch: {
			// Resetting amount if more than can be traded is selected
			buySellAmount: function() {
				if (buySellModal.isBuying) {
					if (buySellModal.buySellAmount > buySellModal.stock.open_shares) {
						buySellModal.buySellAmount = buySellModal.stock.open_shares;
					}
					// determine users cash and limit on purchase cost
					let cash = buySellModal.user.wallet;
					let purchase_val = buySellModal.stock.current_price * buySellModal.buySellAmount;
					if (purchase_val > cash) {
						buySellModal.buySellAmount = Math.floor(cash / buySellModal.stock.current_price); 
					}
				} else {
					//determine current users holdings
					let stock = vm_dash_tab.currUserStocks.filter(d => d.stock_id == buySellModal.stock_uuid)[0];
					if (stock !== undefined) {
	    				if (buySellModal.buySellAmount > stock.amount) {
	    					buySellModal.buySellAmount = stock.amount;
	    				}
					}
				}
			}
		}
	});

	// Vue object for the buy and sell modal
	var genericModal = new Vue({
		el: '#generic-modal--container',
		data: {
			showModal: false,
			// investor_uuid: '',
			investor_name: 'DieselBeaver',
		},
		methods: {
			toPrice: formatPrice,
			addAmount: function(amt) {
				buySellModal.buySellAmount += amt; 
			},
			clearAmount: function() {
				buySellModal.buySellAmount = 0;
			},
			determineMax: function() {
				if (buySellModal.isBuying) {
					buySellModal.buySellAmount = buySellModal.stock.open_shares; 
				} else {
					//determine current users holdings
					let stock = vm_dash_tab.currUserStocks.filter(d => d.stock_id === buySellModal.stock_uuid)[0];
					if (stock !== undefined) {
						buySellModal.buySellAmount = stock.amount;
					} else {
						buySellModal.buySellAmount = 0;
					}
				}
			},
			setIsBuying: function(bool) {
				// Change buying or selling
				buySellModal.isBuying = bool;

				// Set styling
				if (buySellModal.isBuying) {
		        	$('#calc-btn-buy').addClass("fill");
		        	$('#calc-btn-sell').removeClass("fill");
				} else {
		        	$('#calc-btn-sell').addClass("fill");
		        	$('#calc-btn-buy').removeClass("fill");
				}
			},
			submitTrade: function() {
				// Change amount depending on buy/sell
				if (!buySellModal.isBuying) {
					buySellModal.buySellAmount *= -1;
				}
				sendTrade();
				toggleModal();
			},
			closeModal: function(){
				toggleGenericModal();
				// genericModal.investor_uuid = '';
				// genericModal.investor_name = '';
				genericModal.showModal = false;
				
			}
		},
		computed: {
			stock: function() {

				var clickedStock = Object.values(vm_stocks.stocks).filter(d => d.uuid === buySellModal.stock_uuid)[0];
				return clickedStock;
			},
			user: function() {
				var currUserUUID = sessionStorage.getItem('uuid');
				if (vm_users.users[currUserUUID] !== undefined) {
	    			var currUserFolioUUID = vm_users.users[currUserUUID].portfolio_uuid;
					if (vm_portfolios.portfolios[currUserFolioUUID] !== undefined) {
		    			var folio = vm_portfolios.portfolios[currUserFolioUUID];
						folio.investments = folio.net_worth - folio.wallet;
		    			return folio;
	    			}
	    		}
	    		return {};
			}
		},
		watch: {
			// Resetting amount if more than can be traded is selected
			buySellAmount: function() {
				if (buySellModal.isBuying) {
					if (buySellModal.buySellAmount > buySellModal.stock.open_shares) {
						buySellModal.buySellAmount = buySellModal.stock.open_shares;
					}
					// determine users cash and limit on purchase cost
					let cash = buySellModal.user.wallet;
					let purchase_val = buySellModal.stock.current_price * buySellModal.buySellAmount;
					if (purchase_val > cash) {
						buySellModal.buySellAmount = Math.floor(cash / buySellModal.stock.current_price); 
					}
				} else {
					//determine current users holdings
					let stock = vm_dash_tab.currUserStocks.filter(d => d.stock_id == buySellModal.stock_uuid)[0];
					if (stock !== undefined) {
	    				if (buySellModal.buySellAmount > stock.amount) {
	    					buySellModal.buySellAmount = stock.amount;
	    				}
					}
				}
			}
		}
	});


	var allViews = $('.view');
	var dashboardView = $('#dashboard--view');
	var businessView = $('#business--view');
	var stocksView = $('#stocks--view');
	var investorsView = $('#investors--view');
	var futuresView = $('#futures--view');
	var storeView = $('#store--view');
	var currentViewName = $('#current-view');

	function renderContent(route) {
		switch (route) {
			case 'dashboard':
					allViews.removeClass('active');
					dashboardView.addClass('active');
					currentViewName[0].innerHTML = "Dashboard";
			    break;

			case 'business':
					allViews.removeClass('active');
					businessView.addClass('active');
					console.log(currentViewName)
					currentViewName[0].innerHTML = "Business";
			  	break;

			case 'stocks':
					allViews.removeClass('active');
					stocksView.addClass('active');
					currentViewName[0].innerHTML = "Stocks";
			    break;

			case 'investors':
					allViews.removeClass('active');
					investorsView.addClass('active');
					currentViewName[0].innerHTML = "Investors";
			    break;

			case 'futures':
					allViews.removeClass('active');
					futuresView.addClass('active');
					currentViewName[0].innerHTML = "Futures";
			    break;

			case 'perks':
					allViews.removeClass('active');
					storeView.addClass('active');
					currentViewName[0].innerHTML = "Perks";
			    break;
		}
	}

	// SOUND EFFECTS

	var notification_sound = new Audio();
	notification_sound.src = "assets/sfx_pling.wav";

});