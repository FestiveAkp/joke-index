import {useEffect, useState} from 'react';

function App() {
    const [count, setCount] = useState(0);
    const [data, setData] = useState([]);
    console.log(data);

    useEffect(() => {
        const eventSource = new EventSource('https://great-pots-shop.loca.lt/events?stream=jerma985');
        eventSource.onmessage = e => {
            const message = e.data.split(' ');
            setData(data => [...data, {time: new Date(parseInt(message[0]) * 1000), count: message[1]}]);
        };
        return () => eventSource.close();
    }, []);

    return (
        <>
            <h1>Vite + React</h1>
            <div className='card'>
                <button onClick={() => setCount(count => count + 1)}>count is {count}</button>
                <p>
                    Edit <code>src/App.jsx</code> and save to test HMR
                </p>
            </div>
            <p className='read-the-docs'>Click on the Vite and React logos to learn more</p>
        </>
    );
}

export default App;
