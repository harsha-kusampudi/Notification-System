import React, { useState } from 'react';
import DatePicker from 'react-datepicker';
import moment from 'moment';
import axios from 'axios';
import 'react-datepicker/dist/react-datepicker.css';

const NotificationSystem = () => {
  const [message, setMessage] = useState('');
  const [selectedDate, setSelectedDate] = useState(new Date());

  const handleMessageChange = (event) => {
    setMessage(event.target.value);
  };

  const handleDateChange = (date) => {
    setSelectedDate(date);
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    const formattedDate = moment(selectedDate).format('YYYY-MM-DD HH:mm:ss');
    
    const notificationData = {
      message: message,
      timestamp: formattedDate
    };

    try {
      const response = await axios.post('http://localhost:8081/schedule', notificationData);
      console.log('Notification sent:', response.data);
      // Clear the form after successful submission
      setMessage('');
      setSelectedDate(new Date());
      alert('Notification sent successfully!');
    } catch (error) {
      console.error('Error sending notification:', error);
      alert('Failed to send notification. Please try again.');
    }
  };

  return (
    <div style={{ maxWidth: '500px', margin: '0 auto', padding: '20px' }}>
      <h1>Notification System</h1>
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '15px' }}>
          <label htmlFor="message" style={{ display: 'block', marginBottom: '5px' }}>Message:</label>
          <textarea
            id="message"
            value={message}
            onChange={handleMessageChange}
            required
            style={{ width: '100%', height: '100px', padding: '5px' }}
          />
        </div>
        <div style={{ marginBottom: '15px' }}>
          <label htmlFor="datetime" style={{ display: 'block', marginBottom: '5px' }}>Date and Time:</label>
          <DatePicker
            id="datetime"
            selected={selectedDate}
            onChange={handleDateChange}
            showTimeSelect
            timeFormat="HH:mm:ss"
            timeIntervals={1}
            timeCaption="Time"
            dateFormat="yyyy-MM-dd HH:mm:ss"
            required
            style={{ width: '100%', padding: '5px' }}
          />
        </div>
        <button type="submit" style={{ padding: '10px 15px', backgroundColor: '#4CAF50', color: 'white', border: 'none', cursor: 'pointer' }}>
          Schedule Notification
        </button>
      </form>
    </div>
  );
};

export default NotificationSystem;