import streamlit as st
import pandas as pd
import os
import yaml
from pathlib import Path


def load_config():
    """Load configuration from config.yaml"""
    config_paths = [
        "/app/config.yaml",  # Docker path
        "../config.yaml",    # Relative path for local development
        "config.yaml",       # Current directory
    ]

    for config_path in config_paths:
        if os.path.exists(config_path):
            with open(config_path, 'r') as f:
                return yaml.safe_load(f)

    return None


def get_itinerary_metadata():
    """Load itinerary metadata from config.yaml"""
    config = load_config()
    if not config or 'itineraries' not in config:
        return {}

    # Map output files to itinerary metadata
    metadata = {}
    for itin in config['itineraries']:
        output_file = itin.get('output_file', '')
        metadata[output_file] = {
            'id': itin.get('id', 'unknown'),
            'name': itin.get('name', 'Unnamed Itinerary'),
            'from': itin.get('from', 'Unknown'),
            'to': itin.get('to', 'Unknown'),
        }

    return metadata


def load_commute_time(file):
    df = pd.read_csv(file, names=["datetime", "commute_time"])
    df_dt = pd.to_datetime(df["datetime"], utc=True)
    df_dt = df_dt.dt.tz_convert("US/Eastern")

    df = df[df_dt.dt.minute.isin([0, 30])]
    df["hour_min"] = df_dt.dt.strftime("%H:%M")
    df_weekday_nbr = df_dt.dt.weekday + 1
    df_weekday_name = df_dt.dt.day_name()
    df["weekday"] = df_weekday_nbr.astype(str) + " - " + df_weekday_name
    return df


def get_average_commute_time(df: pd.DataFrame):
    avg_time = df.groupby("hour_min")["commute_time"].mean()
    avg_time = avg_time.reset_index()
    return avg_time


def get_average_commute_time_daywise(df: pd.DataFrame):
    avg_time_daywise = df.groupby(["weekday", "hour_min"])["commute_time"].mean()
    avg_time_daywise = avg_time_daywise.reset_index()
    return avg_time_daywise


def get_all_csv_files():
    """Get all CSV files from the data directory"""
    data_dir = "data"
    if not os.path.exists(data_dir):
        return []

    csv_files = [f for f in os.listdir(data_dir) if f.endswith('.csv')]
    return csv_files


def order_csv_files_by_config(csv_files, metadata):
    """Order CSV files by itinerary ID from config"""
    if not metadata:
        return sorted(csv_files)

    # Create a list of (csv_file, sort_key) tuples
    files_with_keys = []
    for csv_file in csv_files:
        file_meta = metadata.get(csv_file, {})
        itin_id = file_meta.get('id', csv_file)
        files_with_keys.append((csv_file, itin_id))

    # Sort by itinerary ID
    files_with_keys.sort(key=lambda x: x[1])

    # Return just the filenames
    return [f[0] for f in files_with_keys]


def display_itinerary(csv_file, metadata):
    """Display a single itinerary's data"""
    file_path = f"data/{csv_file}"

    # Get metadata for this file
    file_metadata = metadata.get(csv_file, {
        'id': csv_file.replace('.csv', ''),
        'name': csv_file.replace('.csv', '').replace('-', ' ').title(),
        'from': 'Unknown',
        'to': 'Unknown',
    })

    # Display header with metadata
    st.subheader(f"📍 {file_metadata['name']}")

    col1, col2 = st.columns(2)
    with col1:
        st.caption(f"**From:** {file_metadata['from']}")
    with col2:
        st.caption(f"**To:** {file_metadata['to']}")

    st.divider()

    # Load and display data
    try:
        df = load_commute_time(file_path)

        if len(df) == 0:
            st.warning("No data available for this itinerary yet.")
            return

        # Average commute time
        st.markdown("#### Average commute time")
        adf = get_average_commute_time(df)
        adf["+/-"] = df.groupby("hour_min")["commute_time"].std().values
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            color="+/-",
        )

        # Statistics
        col1, col2, col3 = st.columns(3)
        with col1:
            st.metric("Average", f"{df['commute_time'].mean():.1f} min")
        with col2:
            st.metric("Min", f"{df['commute_time'].min():.1f} min")
        with col3:
            st.metric("Max", f"{df['commute_time'].max():.1f} min")

        # Day-wise breakdown
        st.markdown("#### Day-wise average commute time")
        adf = get_average_commute_time_daywise(df)
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            stack=False,
            color="weekday",
        )

        # Show raw data option
        with st.expander(f"📊 Show raw data ({len(df)} records)"):
            st.dataframe(df, use_container_width=True)

    except FileNotFoundError:
        st.error(f"Data file not found: {csv_file}")
    except Exception as e:
        st.error(f"Error loading data: {str(e)}")


# Run the app
if __name__ == "__main__":
    st.set_page_config(
        page_title="Commute Time Dashboard",
        page_icon="🚗",
        layout="wide"
    )

    st.title("🚗 Commute Time Dashboard")

    # Load metadata from config
    metadata = get_itinerary_metadata()

    # Show config status
    if metadata:
        st.success(f"✅ Loaded {len(metadata)} itineraries from config.yaml")
    else:
        st.warning("⚠️ Could not load config.yaml. Displaying files from data directory.")

    # Get all CSV files and order them by itinerary ID
    csv_files = get_all_csv_files()
    csv_files = order_csv_files_by_config(csv_files, metadata)

    if not csv_files:
        st.error("No data files found in the data directory.")
        st.stop()

    # Create tabs for each CSV file
    tab_labels = []
    for csv_file in csv_files:
        file_meta = metadata.get(csv_file, {})
        label = file_meta.get('name', csv_file.replace('.csv', ''))
        tab_labels.append(label)

    tabs = st.tabs(tab_labels)

    for i, csv_file in enumerate(csv_files):
        with tabs[i]:
            display_itinerary(csv_file, metadata)
