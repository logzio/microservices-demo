# High level observability demo
In this demo scenario you assume the role of a site reliability engineer (SRE) at the Sock Shop responding to complaints of users receiving errors on the front end and as a result not being able to place orders in the online shop.  You're tasked with resolving the problem as soon as possible so that normal business service can be resumed.

## Establishing the problem
In this stage we will be simulating the business informing the SRE team that something has gone wrong with an alert.
1. Navigate to the [#sock-shop-bizops](https://logzio-demo.slack.com/archives/C027NTKURGR) channel in the [Logz.io Alerts](https://logzio-demo.slack.com/)  Slack workspace
2. Draw the audience's attention to the most recent alert entitled **[Alerting] No new incoming orders (Sock Shop)**
    > This alert tells us that no new orders have been placed in the past 5 minutes, which is highly unusual for the sock shop.  It is simulating customers complaining to support/social media outlets about not being able to place an order on the site.
3. Click on the alert title to be taken to the Logz.io app

## Initial investigation
Here we need to find out where in our Sock Shop kubernetes deployment the problem lies.  With multiple services in play it can usually be challenging to isolate the problem down to a particular service or container.
1. You're presented with the metrics visualization showing the order rate running up to the alert being triggered.  
2. Hover over the first marker in the series - this is demonstrating our **deployment markers** feature and tells us that __Tomer__ deployed a new version of the payments service.  The order rate then dropped soon after this, and the alert was triggered.
3. Click the __Go back__ arrow in the top left hand corner of the visualization to zoom out to the Sock Shop metrics dashboard.
    > As well as demonstrating business metrics such as **daily order volume** and **new order rate** this dashboard shows us the number of **queries per second** (QPS) and **latency** for each of our services.  These are typical metrics which demonstrate how well a microservice is performing.
4. Look at the **Daily Order Volume** visualization, and notice that it has levelled off, showing that no new orders have been logged.
5. From the left hand (QPS) column, see that there has been a drop off in successful queries on all of the services.  This means that successful user interactions with the site has dropped.
6. On the right hand column (latency), see that there is high latency on the **Frontend, Orders and Payment services**, now being mesaured in seconds rather than ms.  This tells us that our site has become unresponsive for our customers.
7. Finally, in the QPS column notice that the error rate, represented by the yellow 4xx/5xx line, for the **Orders** and **Payment** services is increasing.  This tells us that there is a problem in one or both of these services.
     > 4xx/5xx represent [HTTP error codes](https://en.wikipedia.org/wiki/List_of_HTTP_status_codes) (e.g. 404 not found).  If everything is working fine in a service then the normal HTTP status is in the 2xx or 3xx range.
8. We know there is a problem in one of these services, but need to invstigate the transaction traces to be able to isolate the problem further.  Click the **Kibana Traces Dashboard** link in the top right hand corner of the dashboard.

## Isolating the broken service
Our traces dashboard gives us a detailed breakdown of how our microservices are performing.  The data is all based around trace information coming in from OpenTelemetry auto-instrumentation in our Java and Go services.
1. As you scoll down the dashboard, notice that the **Duration Percentiles** and **Average Duration Per Service** visualizations both show an increase in time, reflecting what we saw in our metrics dashboard earlier.
2.  Scroll down to the **Errors Summary** area in the dashboard.  Point out that there is a spike in errors on both the **orders** and **payment** service.
    > There are already errors being reported in the **orders** service but they spike after the deployment happens. These errors are related to declined payments and aren't part of the issue we are troubleshooting here.
3. In the **#of Errors per Service** list, we have a link to the last trace for each service which has logged an error.  Click on the **Last TraceID** for the **orders** service.  This will take us to the Jaeger Tracing view for that individual trace.

## Using tracing to troubleshoot an individual transaction
Now we have isolated an individual transaction which has errors it time to look at where in our microservice architecture things are breaking.
1. Point out that the majority of spans in the **orders** and **carts** service are returning successfully in milliseconds, indicating that these spans are performing as expected.  However, when **orders** is making an HTTP POST request, the span duration is seconds rather than ms, indicating that something is taking longer than usual and/or timing out.
2. Click on the span for **orders HTTP POST** which is marked with an error symbol.  Looking at the tags which have been automatically captured using the OpenTelemetry Java Agent, highlight the following important pieces of information:
    - _http.status_code_ is 500, capturing an error in the service which we have seen in the metrics dashboard earlier
    - _http.url_ is http://payment/paymentAuth indicating that the orders service is calling the service method which authorizes payments
3. Click on the child span for **payment payment** (not the deeper **authorize payment** span).  Point out that the _http.route_ and _http.status_code_ match the **orders** parent span, telling us that the request is getting from the orders to the payments service, indicating that the **payments** service is the broken one.
4. Expand the **Logs** section in this span.  It just shows the number of read/write bytes which has automatically been captured through auto-instrumentation.  This information does not really help us in troubleshooting this issue.
5. To further isolate the problem, we need to look at the application logs for the payments service related to this trace.  Click the **View in kibana** link in the top left hand corner of the trace (just under the **JAEGER UI**) text.

## Searching the application logs for errors
In Kibana we can see all of the log lines from all of the services involved, some of which have been automatically captured through OpenTelemetry auto-instrumentation, and some of which have been manually annotated with the trace ID.
1. Click the **Exceptions** tab and expand the first exception in the list.  Point out that this has been automatically picked up as an exception by Logz.io's machine learning analysis engine.
2. Looking at the error message, and the **process.servicename** field we see that this is related to the orders service, which a _symptom_ but not the _root cause_.  We need to look at log items just related to the **payments** service.
3.  Go back to the **Logs** tab and click **Add filter** in the top left hand corner of Kibana (just below the search box).  Select (or type) **kubernetes.container_name** as the Field and **payment** as the Value.
4.  Now we can see the real error which has been captured - the error message is _Payment Gateway Unavailable: dev.sock-shop.payments.com_
    > This kind of problem occurs when application developers hard code variables in their code, which in this case works in their _DEV_ environment, but when the application is built and deployed in _PROD_ is breaks as the production environment is isolated and can't access a service in _DEV_.  In our architecure the payments service is calling out to a third-party payment gateway which is simulated in the _DEV_ environment at dev.sock-shop.payments.com but we have already deployed that configuration to _PROD_ hence payments aren't authorized when customers are trying to place an order.
5. Expand the document and click the **Filter for value** icon which appears when you hover over the error field.
6. Remove the traceID frome the Kibana search box and click to **Refresh** the screen.
7. Now Kibana is displaying all of the log lines related to the **payment** service which also contain this error.  To be more proactive in future and get notifications as soon as this issue arises then we would normally create an alert (don't do this as we have already done it in the [#sock-shop-sre](https://logzio-demo.slack.com/archives/C0274QZ1XEK) Slack channel

 ## Wrapping it all up
The fix for this error was to roll back the payments service to a known good deployment and that's what we did.  Let's take a look at the health of the system after we did that.
1. In the [#sock-shop-bizops](https://logzio-demo.slack.com/archives/C027NTKURGR) Slack channel click on the latest **OK** alert.  This takes us to the metrics visualization associated with the alert (New order volume)
2. Zoom out to the dashboard by clicking the link in the top left hand corner of the visualization.  Scroll down to the **Orders QPS** visualization and highlight the second deployment marker, indicating that Asaf deployed version 0.4.3 of the payments service.
3. As the error rate for **orders** and **payment** drops after the deployment, we can be confident that those services are now operating normally.
4. Finally, scroll back up to the top and see the **Order volume** and **rate** starting to increase again, indicating that customers are now able to place orders.  Issue resolved!
5. Optionally, navigate to the [#sock-shop-sre](https://logzio-demo.slack.com/archives/C0274QZ1XEK) Slack channel and show the new alerts, allowing us to be more proactive in resolving this issue if it occurs again.
