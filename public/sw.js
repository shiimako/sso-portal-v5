self.addEventListener('push', function(event) {
  let data = { title: "Portal Notification", body: "Check dashboard", url: "/" };
  
  if (event.data) {
    data = event.data.json();
  }

  const options = {
    body: data.body,
    icon: data.icon || '/public/static/Logo-PNC.png',
    badge: '/public/static/Logo-PNC.png',
    vibrate: [100, 50, 100],
    data: {
      url: data.url
    }
  };

  event.waitUntil(
    self.registration.showNotification(data.title, options)
  );
});

self.addEventListener('notificationclick', function(event) {
  event.notification.close();
  event.waitUntil(
    clients.openWindow(event.notification.data.url)
  );
});