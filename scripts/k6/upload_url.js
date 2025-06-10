import http from 'k6/http';
import {
    check
} from 'k6';

export const options = {
    vus: 25,
    duration: "20s",
    iterations: 50_000,
};

export default function() {
    try {
        // Generate random URL for testing
        const url = `http://example.com/${Math.random()}`;

        // Send POST request to shorten URL
        const response = http.post(
            "http://127.0.0.1:8080/api/shorten",
            JSON.stringify({
                url: url
            }), {
                headers: {
                    "Content-Type": "application/json"
                }
            }
        );

        // Verify the response
        check(response, {
            "status is 200": (r) => r.status === 201,
            "response has shortened URL": (r) => {
                const body = JSON.parse(r.body);
                return body.hasOwnProperty("result");
            },
        });
    } catch (error) {
        console.error('Request failed: ' + error);
    }
}