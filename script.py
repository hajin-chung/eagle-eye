from PIL import Image
import numpy as np
import shutil
import os
import subprocess
import requests
import sys

frames_path = "./frames"
video_path = "./in.mp4"

def download_reels(url):
    print("[*] fetching video_url")
    api_url = f"https://instagram-videos.vercel.app/api/video?url={url}"
    r = requests.get(api_url)
    data = r.json()
    video_url = data['data']['videoUrl']
    print(f"[*] video_url = {video_url}")
    print("[*] downloading video")
    with requests.get(video_url, stream=True) as r:
        r.raise_for_status()
        with open(video_path, 'wb') as f:
            for chunk in r.iter_content(chunk_size=8192):
                f.write(chunk)
    print("[*] success")


def extract_frames():
    print("[*] extracting frames from video")
    subprocess.run(["ffmpeg", "-i", video_path, "./frames/%03d.bmp"])


def read_img(filename):
    pic = Image.open(os.path.join(frames_path, filename))
    return pic 


def capture_frame(id):
    frames = [
        f for f in os.listdir(frames_path) 
            if os.path.isfile(os.path.join(frames_path, f))]
    frames.sort()

    n = len(frames)

    prev = read_img(frames[0])
    max_mse = 0
    target_idx = 0
    for i in range(1, n):
        curr_file = frames[i]
        curr = read_img(curr_file)

        if curr.height != prev.height or curr.width != prev.width:
            curr = curr.resize((prev.width, prev.height))

        mse = ((np.array(prev) - np.array(curr))**2).mean()
        if max_mse < mse:
            max_mse = mse
            target_idx = i

        prev = curr

    print(target_idx, n)
    shutil.copy(os.path.join(frames_path, frames[target_idx-1]), f"./result/{id}-1.bmp")
    shutil.copy(os.path.join(frames_path, frames[target_idx]), f"./result/{id}-2.bmp")


def main(argv):
    if len(argv) < 2:
        sys.exit("need more args")

    url = argv[1]
    id = argv[2]
    download_reels(url)
    extract_frames()
    capture_frame(id)


if __name__ == "__main__":
    main(sys.argv)
